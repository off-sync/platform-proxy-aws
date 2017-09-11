package startproxy

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/off-sync/platform-proxy-app/interfaces"
)

type proxy struct {
	// context
	ctx context.Context
	wg  *sync.WaitGroup

	// logging
	logger interfaces.Logger

	// configuration
	frontendRepository interfaces.FrontendRepository
	serviceRepository  interfaces.ServiceRepository
	pollingDuration    time.Duration

	// request handling
	webServer       interfaces.WebServer
	secureWebServer interfaces.SecureWebServer
	loadBalancer    interfaces.LoadBalancer

	// internal state
	serviceHandlers map[string]http.Handler
	frontendConfigs map[string]*frontendConfig
}

type frontendConfig struct {
	serviceName string

	// url is needed for deleting the configured routes from the web servers
	url *url.URL

	// isSecure is needed for deleting the configured routes from the web servers
	isSecure bool
}

func newProxy(
	ctx context.Context,
	wg *sync.WaitGroup,
	logger interfaces.Logger,
	serviceRepository interfaces.ServiceRepository,
	frontendRepository interfaces.FrontendRepository,
	pollingDuration time.Duration,
	webServer interfaces.WebServer,
	secureWebServer interfaces.SecureWebServer,
	loadBalancer interfaces.LoadBalancer) *proxy {

	return &proxy{
		ctx:                ctx,
		wg:                 wg,
		logger:             logger,
		serviceRepository:  serviceRepository,
		frontendRepository: frontendRepository,
		pollingDuration:    pollingDuration,
		webServer:          webServer,
		secureWebServer:    secureWebServer,
		loadBalancer:       loadBalancer,
		serviceHandlers:    make(map[string]http.Handler),
		frontendConfigs:    make(map[string]*frontendConfig),
	}
}

func (p *proxy) run() {
	// configure all services and frontends
	p.configure()

	// subscribe to service events
	serviceEvents := make(<-chan interfaces.ServiceEvent)

	if w, ok := p.serviceRepository.(interfaces.ServiceWatcher); ok {
		p.logger.Info("subscribing to service watcher")

		serviceEvents = w.Subscribe()
	}

	// subscribe to frontend events
	frontendEvents := make(<-chan interfaces.FrontendEvent)

	if w, ok := p.frontendRepository.(interfaces.FrontendWatcher); ok {
		p.logger.Info("subscribing to frontend watcher")

		frontendEvents = w.Subscribe()
	}

	// create polling ticker
	pollTicker := time.NewTicker(p.pollingDuration)

	p.wg.Add(1)

	for {
		select {
		// respond to the context closing
		case <-p.ctx.Done():
			pollTicker.Stop()
			p.wg.Done()

			p.logger.Info("context is done: returning")
			return

			// respond to polling events
		case <-pollTicker.C:
			p.logger.Info("polling configuration")
			p.configure()
			break

			// respond to service events
		case serviceEvent := <-serviceEvents:
			p.logger.
				WithField("name", serviceEvent.Name).
				Info("received service event")

			p.configureService(serviceEvent.Name)

			break

			// respond to frontend events
		case frontendEvent := <-frontendEvents:
			p.logger.
				WithField("name", frontendEvent.Name).
				Info("received frontend event")

			p.configureFrontend(frontendEvent.Name)

			break
		}
	}
}

func (p *proxy) configure() {
	// configure services first to create the required handlers
	services, err := p.serviceRepository.ListServices()
	if err != nil {
		p.logger.
			WithError(err).
			Error("listing services")
	} else {
		currentServices := make(map[string]bool)

		for _, service := range services {
			p.configureService(service)

			currentServices[service] = true
		}

		for service := range p.serviceHandlers {
			if _, found := currentServices[service]; !found {
				// call configure service to have it removed
				p.configureService(service)
			}
		}
	}

	// configure frontends
	frontends, err := p.frontendRepository.ListFrontends()
	if err != nil {
		p.logger.
			WithError(err).
			Error("listing frontends")
	} else {
		currentFrontends := make(map[string]bool)

		for _, frontend := range frontends {
			p.configureFrontend(frontend)

			currentFrontends[frontend] = true
		}

		for frontend := range p.frontendConfigs {
			if _, found := currentFrontends[frontend]; !found {
				// call configure frontend to have it removed
				p.configureFrontend(frontend)
			}
		}
	}
}

func (p *proxy) getServiceHandler(serviceName string) http.Handler {
	handler, found := p.serviceHandlers[serviceName]
	if !found {
		return http.NotFoundHandler()
	}

	return handler
}

func (p *proxy) configureService(name string) {
	// describe service
	service, err := p.serviceRepository.DescribeService(name)
	if err != nil {
		// check if error means that service does not exists
		if err == interfaces.ErrUnknownService {
			// check if service was configured previously
			if _, found := p.serviceHandlers[name]; !found {
				return
			}

			p.logger.
				WithField("name", name).
				Debug("deleting service")

			// delete service handler mapping
			delete(p.serviceHandlers, name)

			// reconfigure linked frontends
			for frontendName, frontendConfig := range p.frontendConfigs {
				if frontendConfig.serviceName != name {
					continue
				}

				p.configureFrontend(frontendName)
			}

			// delete load balancer service
			p.loadBalancer.DeleteService(name)

			return
		}

		p.logger.
			WithError(err).
			WithField("name", name).
			Error("describing service")

		return
	}

	p.logger.
		WithField("name", service.Name).
		WithField("servers", service.Servers).
		Debug("upserting service")

	handler, err := p.loadBalancer.UpsertService(service.Name, service.Servers...)
	if err != nil {
		p.logger.
			WithError(err).
			WithField("name", name).
			WithField("servers", service.Servers).
			Error("upserting service")

		// set the service handler to return an internal server error on each
		// request
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Service not configured", http.StatusInternalServerError)
		})
	}

	// upsert service handler mapping
	p.serviceHandlers[service.Name] = handler
}

func (p *proxy) configureFrontend(name string) {
	// get frontend from repository
	frontend, err := p.frontendRepository.DescribeFrontend(name)
	if err != nil {
		// check if error means that frontend does not exist
		if err == interfaces.ErrUnknownFrontend {
			// check if frontend was configured previously
			frontendConfig, found := p.frontendConfigs[name]
			if !found {
				return
			}

			p.logger.
				WithField("name", name).
				Debug("deleting frontend")

			// delete frontend config
			delete(p.frontendConfigs, name)

			// delete web server routes
			if frontendConfig.isSecure {
				p.secureWebServer.DeleteRoute(frontendConfig.url)

				httpURL := &url.URL{}
				*httpURL = *frontendConfig.url
				httpURL.Scheme = "http"

				p.webServer.DeleteRoute(httpURL)
			} else {
				p.webServer.DeleteRoute(frontendConfig.url)
			}

			return
		}

		p.logger.
			WithError(err).
			WithField("name", name).
			Error("describing frontend")

		return
	}

	p.logger.
		WithField("name", frontend.Name).
		WithField("url", frontend.URL).
		WithField("service_name", frontend.ServiceName).
		Debug("configuring frontend")

	if frontend.Certificate != nil {
		// configure HTTPS
		err := p.secureWebServer.UpsertCertificate(
			frontend.URL.Host,
			frontend.Certificate)
		if err != nil {
			p.logger.
				WithError(err).
				WithField("host", frontend.URL.Host).
				Error("upserting certificate")
		}

		err = p.secureWebServer.UpsertRoute(
			frontend.URL,
			p.getServiceHandler(frontend.ServiceName))
		if err != nil {
			p.logger.
				WithError(err).
				WithField("url", frontend.URL).
				Error("upserting route")
		}

		// configure HTTP redirect
		httpURL := &url.URL{}
		*httpURL = *frontend.URL
		httpURL.Scheme = "http"

		err = p.webServer.UpsertRoute(httpURL,
			http.RedirectHandler(
				frontend.URL.String(),
				http.StatusMovedPermanently))
		if err != nil {
			p.logger.
				WithError(err).
				WithField("url", frontend.URL).
				Error("upserting route")
		}
	} else {
		// configure HTTP
		err := p.webServer.UpsertRoute(
			frontend.URL,
			p.getServiceHandler(frontend.ServiceName))
		if err != nil {
			p.logger.
				WithError(err).
				WithField("url", frontend.URL).
				Error("upserting route")
		}
	}

	// upsert service handler mapping
	p.frontendConfigs[name] = &frontendConfig{
		serviceName: frontend.ServiceName,
		url:         frontend.URL,
		isSecure:    frontend.Certificate != nil,
	}
}
