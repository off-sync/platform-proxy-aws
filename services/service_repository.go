package services

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/off-sync/platform-proxy-domain/services"
)

// ServiceRepository implements the ServiceRepository interface using an AWS ECS
// cluster as its backend.
type ServiceRepository struct {
	// AWS ECS API
	api interfaces.AwsEcsAPI

	// Configuration
	serverContainerName string
	dockerLabelPort     string
	defaultPort         int
}

// Default values for the ServiceRepository struct.
const (
	DefaultServerContainerName = "server"
	DefaultDockerLabelPort     = "com.off-sync.platform.proxy.port"
	DefaultDefaultPort         = 8080
)

// ServiceRepositoryOption defines the type used to further configure a
// ServiceRepository.
type ServiceRepositoryOption func(*ServiceRepository) error

// NewServiceRepository creates a new service repository based on the provided
// AWS ECS API.
func NewServiceRepository(api interfaces.AwsEcsAPI, options ...ServiceRepositoryOption) (*ServiceRepository, error) {
	r := &ServiceRepository{
		api:                 api,
		serverContainerName: DefaultServerContainerName,
		dockerLabelPort:     DefaultDockerLabelPort,
		defaultPort:         DefaultDefaultPort,
	}

	for _, opt := range options {
		err := opt(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// WithServerContainerName configures a service repository with the provided
// server container name.
func WithServerContainerName(name string) ServiceRepositoryOption {
	return func(r *ServiceRepository) error {
		r.serverContainerName = name
		return nil
	}
}

// WithDockerLabelPort configures a service repository with the provided
// docker label for the port.
func WithDockerLabelPort(label string) ServiceRepositoryOption {
	return func(r *ServiceRepository) error {
		r.dockerLabelPort = label
		return nil
	}
}

// WithDefaultPort configures a service repository with the provided
// default port.
func WithDefaultPort(port int) ServiceRepositoryOption {
	return func(r *ServiceRepository) error {
		r.defaultPort = port
		return nil
	}
}

// ListServices returns all service names contained in this repository.
func (r *ServiceRepository) ListServices() ([]string, error) {
	return r.api.ListServices()
}

// DescribeService returns the service with the specified name. If no service
// exists with that name an ErrUnknownService is returned.
func (r *ServiceRepository) DescribeService(name string) (*services.Service, error) {
	service, err := r.api.DescribeService(name)
	if err != nil {
		return nil, err
	}

	serverURL, err := r.getTaskDefinitionServerURL(aws.StringValue(service.TaskDefinition))
	if err != nil {
		return nil, err
	}

	return services.NewService(name, serverURL)
}

func (r *ServiceRepository) getTaskDefinitionServerURL(taskDefArn string) (string, error) {
	tdef, err := r.api.DescribeTaskDefinition(taskDefArn)
	if err != nil {
		return "", err
	}

	for _, cdef := range tdef.ContainerDefinitions {
		if aws.StringValue(cdef.Name) != r.serverContainerName {
			// not the server
			continue
		}

		port := r.defaultPort

		portLabel, found := cdef.DockerLabels[r.dockerLabelPort]
		if found {
			port, err = strconv.Atoi(*portLabel)
			if err != nil {
				return "", fmt.Errorf("invalid port: %s", *portLabel)
			}
		}

		return fmt.Sprintf("http://%s:%d", *cdef.Hostname, port), nil
	}

	return "", fmt.Errorf("no server container found for task definition: %s", taskDefArn)
}
