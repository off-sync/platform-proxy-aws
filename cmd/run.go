// Copyright (c) 2017 off-sync
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial Addrions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/off-sync/platform-proxy-app/infra/logging"
	"github.com/off-sync/platform-proxy-app/proxies/cmd/startproxy"
	"github.com/off-sync/platform-proxy-aws/frontends"
	"github.com/off-sync/platform-proxy-aws/infra"
	"github.com/off-sync/platform-proxy-aws/loadbalancers"
	"github.com/off-sync/platform-proxy-aws/services"
	"github.com/off-sync/platform-proxy-aws/webservers"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the Off-Sync.com Platform Proxy for Amazon Web Services",
	Long:  ``,
	Run:   run,
}

// Flags, configuration keys and defaults.
const (
	KeyPrefix = "run."

	FlagPollingDuration    = "polling-duration"
	KeyPollingDuration     = KeyPrefix + "pollingDuration"
	DefaultPollingDuration = 300

	FlagAddr    = "addr"
	KeyAddr     = KeyPrefix + "addr"
	DefaultAddr = ":80"

	FlagSecureAddr    = "secure-addr"
	KeySecureAddr     = KeyPrefix + "secureAddr"
	DefaultSecureAddr = ":443"
)

func init() {
	RootCmd.AddCommand(runCmd)

	// polling duration flag and configuration
	runCmd.Flags().Int32P(FlagPollingDuration, "d", DefaultPollingDuration, "Polling duration in seconds")

	viper.SetDefault(KeyPollingDuration, DefaultPollingDuration)
	viper.BindPFlag(KeyPollingDuration, runCmd.Flags().Lookup(FlagPollingDuration))

	// HTTP Addr flag and configuration
	runCmd.Flags().StringP(FlagAddr, "a", DefaultAddr, "Address used by the Web Server")

	viper.SetDefault(KeyAddr, DefaultAddr)
	viper.BindPFlag(KeyAddr, runCmd.Flags().Lookup(FlagAddr))

	// HTTPS Addr flag and configuration
	runCmd.Flags().StringP(FlagSecureAddr, "s", DefaultSecureAddr, "Address used by the Secure Web Server")

	viper.SetDefault(KeySecureAddr, DefaultSecureAddr)
	viper.BindPFlag(KeySecureAddr, runCmd.Flags().Lookup(FlagSecureAddr))
}

func run(cmd *cobra.Command, args []string) {
	ecsAPI, err := infra.NewAwsEcsSdkFromConfig()
	if err != nil {
		logger.
			WithError(err).
			Fatal("creating AWS ECS API")

		return
	}

	dynAPI, err := infra.NewAwsDynamoDBSdkFromConfig()
	if err != nil {
		logger.
			WithError(err).
			Fatal("creating AWS DynamoDB API")

		return
	}

	serviceRepository, err := services.NewServiceRepository(ecsAPI)
	if err != nil {
		logger.
			WithError(err).
			Fatal("creating service repository")

		return
	}

	svcs, err := serviceRepository.ListServices()
	if err != nil {
		logger.WithError(err).Error("listing services")
	} else {
		for _, svc := range svcs {
			logger.WithField("name", svc).Info("found service")
		}
	}

	frontendRepository, err := frontends.NewFrontendRepository(dynAPI, viper.GetString("dyndbFrontendsTable"))
	if err != nil {
		logger.
			WithError(err).
			Fatal("creating frontend repository")

		return
	}

	frontends, err := frontendRepository.ListFrontends()
	if err != nil {
		logger.WithError(err).Error("listing frontends")
	} else {
		for _, frontend := range frontends {
			logger.WithField("name", frontend).Info("found frontend")
		}
	}

	startProxyCmd, err := startproxy.NewCommand(
		serviceRepository,
		frontendRepository,
		logging.NewLogrusLogger(logger))
	if err != nil {
		logger.WithError(err).Fatal("creating start proxy command")

		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	err = startProxyCmd.Execute(&startproxy.Model{
		Ctx:             ctx,
		WaitGroup:       wg,
		PollingDuration: time.Duration(viper.GetInt(KeyPollingDuration)) * time.Second,
		LoadBalancer:    &loadbalancers.LoadBalancer{},
		SecureWebServer: webservers.NewSecureWebServer(logging.NewLogrusLogger(logger), viper.GetString(KeySecureAddr)),
		WebServer:       webservers.NewWebServer(logging.NewLogrusLogger(logger), viper.GetString(KeyAddr)),
	})
	if err != nil {
		logger.WithError(err).Fatal("executing start proxy command")
	}

	// wait for SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	logger.Info("received SIGINT: stopping")

	// cancel the context to trigger the cleanup process
	cancel()

	// wait for any processes that were started by the start proxy command
	// to cleanup
	wg.Wait()

	logger.Info("stopped")
}
