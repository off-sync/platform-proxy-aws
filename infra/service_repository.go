package infra

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy-domain/services"
)

// ServiceRepository implements the ServiceRepository interface using an AWS ECS
// cluster as its backend.
type ServiceRepository struct {
	ecsSvc  *ecs.ECS
	cluster *ecs.Cluster

	// Configuration
	serverContainerName string
	dockerLabelPort     string
	defaultPort         int
}

// ServiceRepositoryOption defines the type used to further configure a
// ServiceRepository.
type ServiceRepositoryOption func(*ServiceRepository) error

// NewServiceRepository creates a new service repository based on the provided
// AWS ECS client and cluster name.
func NewServiceRepository(
	ecsSvc *ecs.ECS,
	clusterName string,
	options ...ServiceRepositoryOption) (*ServiceRepository, error) {
	clusters, err := ecsSvc.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	})
	if err != nil {
		return nil, err
	}

	if len(clusters.Failures) > 0 {
		return nil, fmt.Errorf("checking cluster: %s", *clusters.Failures[0].Reason)
	}

	if len(clusters.Clusters) < 1 {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	r := &ServiceRepository{
		ecsSvc:              ecsSvc,
		cluster:             clusters.Clusters[0],
		serverContainerName: "server",
		dockerLabelPort:     "com.off-sync.platform.proxy.port",
		defaultPort:         8080,
	}

	for _, opt := range options {
		err = opt(r)
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
	var serviceNames []string

	err := r.ecsSvc.ListServicesPages(&ecs.ListServicesInput{
		Cluster: r.cluster.ClusterName,
	}, func(output *ecs.ListServicesOutput, lastPage bool) bool {
		serviceNames = append(serviceNames, aws.StringValueSlice(output.ServiceArns)...)
		return true
	})
	if err != nil {
		return nil, err
	}

	return serviceNames, nil
}

// DescribeService returns the service with the specified name. If no service
// exists with that name an ErrUnknownService is returned.
func (r *ServiceRepository) DescribeService(name string) (*services.Service, error) {
	serviceDescription, err := r.ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  r.cluster.ClusterName,
		Services: aws.StringSlice([]string{name}),
	})
	if err != nil {
		return nil, err
	}

	if len(serviceDescription.Services) < 1 {
		return nil, fmt.Errorf("service not found for name: %s", name)
	}

	service := serviceDescription.Services[0]

	serverURL, err := r.getTaskDefinitionServer(service.TaskDefinition)
	if err != nil {
		return nil, err
	}

	return services.NewService(aws.StringValue(service.ServiceName), serverURL)
}

func (r *ServiceRepository) getTaskDefinitionServer(taskDefArn *string) (string, error) {
	tdef, err := r.ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskDefArn,
	})
	if err != nil {
		return "", err
	}

	for _, cdef := range tdef.TaskDefinition.ContainerDefinitions {
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

	return "", nil
}
