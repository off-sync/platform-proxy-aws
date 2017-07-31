package services

import (
	"errors"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"

	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/off-sync/platform-proxy-domain/services"
)

func setUp(t *testing.T, options ...ServiceRepositoryOption) (*ServiceRepository, *interfaces.AwsEcsAPIMock) {
	api := interfaces.NewAwsEcsAPIMock()

	serviceRepository, err := NewServiceRepository(api, options...)

	assert.Nil(t, err)
	assert.NotNil(t, serviceRepository)

	return serviceRepository, api
}

func TestNewServiceRepository(t *testing.T) {
	setUp(t)
}

func TestNewServiceRepositoryWithOptions(t *testing.T) {
	setUp(t,
		WithServerContainerName("name"),
		WithDockerLabelPort("label"),
		WithDefaultPort(1234))
}

func TestNewServiceRepositoryWithFailingOption(t *testing.T) {
	optErr := errors.New("option error")

	api := interfaces.NewAwsEcsAPIMock()

	serviceRepository, err := NewServiceRepository(api, func(*ServiceRepository) error {
		return optErr
	})
	assert.Nil(t, serviceRepository)

	assert.Equal(t, optErr, err)
}

func TestListServices(t *testing.T) {
	r, api := setUp(t)

	api.ServiceNames = []string{"service1", "service2"}

	names, err := r.ListServices()
	assert.Nil(t, err)

	assert.EqualValues(t, []string{"service1", "service2"}, names)
}

func TestDescribeService(t *testing.T) {
	r, api := setUp(t)

	api.Services["service1"] = &ecs.Service{
		TaskDefinition: aws.String("taskDef1"),
	}

	dockerLabels := make(map[string]string)
	dockerLabels[DefaultDockerLabelPort] = "9090"

	api.TaskDefs["taskDef1"] = &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Name: aws.String("not the server"),
			},
			&ecs.ContainerDefinition{
				DockerLabels: aws.StringMap(dockerLabels),
				Name:         aws.String(DefaultServerContainerName),
				Hostname:     aws.String("hostname"),
			},
		},
	}

	svc, err := r.DescribeService("service1")
	assert.Nil(t, err)

	serverURL, err := url.Parse("http://hostname:9090")

	assert.EqualValues(t, &services.Service{
		Name:    "service1",
		Servers: []*url.URL{serverURL},
	}, svc)
}

func TestDescribeServiceShouldReturnErrorWhenAPIFails(t *testing.T) {
	r, api := setUp(t)
	api.FailDescribeService = true

	_, err := r.DescribeService("service1")
	assert.NotNil(t, err)

	r, api = setUp(t)
	api.Services["service1"] = &ecs.Service{TaskDefinition: aws.String("taskDef1")}
	api.FailDescribeTaskDefinition = true

	_, err = r.DescribeService("service1")
	assert.NotNil(t, err)
}

func TestDescribeServiceShouldReturnErrorWhenServerContainerNotFound(t *testing.T) {
	r, api := setUp(t)
	api.Services["service1"] = &ecs.Service{TaskDefinition: aws.String("taskDef1")}
	api.TaskDefs["taskDef1"] = &ecs.TaskDefinition{}

	_, err := r.DescribeService("service1")
	assert.NotNil(t, err)
}

func TestDescribeServiceShouldReturnErrorOnInvalidPort(t *testing.T) {
	r, api := setUp(t)

	api.Services["service1"] = &ecs.Service{
		TaskDefinition: aws.String("taskDef1"),
	}

	dockerLabels := make(map[string]string)
	dockerLabels[DefaultDockerLabelPort] = "abc"

	api.TaskDefs["taskDef1"] = &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				DockerLabels: aws.StringMap(dockerLabels),
				Name:         aws.String(DefaultServerContainerName),
			},
		},
	}

	_, err := r.DescribeService("service1")
	assert.NotNil(t, err)
}
