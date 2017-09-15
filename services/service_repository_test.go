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
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package services

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"

	"github.com/off-sync/platform-proxy-aws/common"
	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/off-sync/platform-proxy-domain/services"
)

func setUpSqsWatcher(t *testing.T) (*common.SqsWatcher, *interfaces.AwsSqsAPIMock) {
	api := interfaces.NewAwsSqsAPIMock()

	sw, err := common.NewSqsWatcher(context.Background(), api, "name")
	assert.Nil(t, err)

	return sw, api
}

func setUp(t *testing.T, options ...ServiceRepositoryOption) (
	*ServiceRepository,
	*interfaces.AwsEcsAPIMock,
	*interfaces.AwsSqsAPIMock) {
	api := interfaces.NewAwsEcsAPIMock()

	sw, sqsAPI := setUpSqsWatcher(t)

	serviceRepository, err := NewServiceRepository(
		api,
		sw,
		options...)

	assert.Nil(t, err)
	assert.NotNil(t, serviceRepository)

	return serviceRepository, api, sqsAPI
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

func TestNewServiceRepositoryWithMissingAPI(t *testing.T) {
	sw, _ := setUpSqsWatcher(t)

	serviceRepository, err := NewServiceRepository(nil, sw)
	assert.Nil(t, serviceRepository)
	assert.Equal(t, ErrMissingAwsEcsAPI, err)
}

func TestNewServiceRepositoryWithMissingSqsWatcher(t *testing.T) {
	api := interfaces.NewAwsEcsAPIMock()

	serviceRepository, err := NewServiceRepository(api, nil)
	assert.Nil(t, serviceRepository)
	assert.Equal(t, ErrMissingSqsWatcher, err)
}

func TestNewServiceRepositoryWithFailingOption(t *testing.T) {
	optErr := errors.New("option error")

	api := interfaces.NewAwsEcsAPIMock()

	sw, _ := setUpSqsWatcher(t)

	serviceRepository, err := NewServiceRepository(api, sw, func(*ServiceRepository) error {
		return optErr
	})
	assert.Nil(t, serviceRepository)

	assert.Equal(t, optErr, err)
}

func TestListServices(t *testing.T) {
	r, api, _ := setUp(t)

	api.ServiceNames = []string{"service1", "service2"}

	names, err := r.ListServices()
	assert.Nil(t, err)

	assert.EqualValues(t, []string{"service1", "service2"}, names)
}

func TestDescribeService(t *testing.T) {
	r, api, _ := setUp(t)

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
		Type:    services.ServiceTypeServer,
		Servers: []*url.URL{serverURL},
	}, svc)
}

func TestDescribeServiceShouldReturnErrorWhenAPIFails(t *testing.T) {
	r, api, _ := setUp(t)
	api.FailDescribeService = true

	_, err := r.DescribeService("service1")
	assert.NotNil(t, err)

	r, api, _ = setUp(t)
	api.Services["service1"] = &ecs.Service{TaskDefinition: aws.String("taskDef1")}
	api.FailDescribeTaskDefinition = true

	_, err = r.DescribeService("service1")
	assert.NotNil(t, err)
}

func TestDescribeServiceShouldReturnSystemTypeWhenDockerLabelIsPresent(t *testing.T) {
	r, api, _ := setUp(t)

	api.Services["service1"] = &ecs.Service{TaskDefinition: aws.String("taskDef1")}

	dockerLabels := make(map[string]string)
	dockerLabels[DefaultDockerLabelServiceType] = "system"

	api.TaskDefs["taskDef1"] = &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				DockerLabels: aws.StringMap(dockerLabels),
			},
		},
	}

	svc, err := r.DescribeService("service1")
	assert.Nil(t, err)
	assert.Equal(t, services.ServiceTypeSystem, svc.Type)
}

func TestDescribeServiceShouldReturnUnsupportedTypeWhenServerContainerNotFound(t *testing.T) {
	r, api, _ := setUp(t)
	api.Services["service1"] = &ecs.Service{TaskDefinition: aws.String("taskDef1")}
	api.TaskDefs["taskDef1"] = &ecs.TaskDefinition{}

	svc, err := r.DescribeService("service1")
	assert.Nil(t, err)
	assert.Equal(t, services.ServiceTypeUnsupported, svc.Type)
}

func TestDescribeServiceShouldReturnErrorOnInvalidPort(t *testing.T) {
	r, api, _ := setUp(t)

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

func TestSubscribe(t *testing.T) {
	r, _, sqsAPI := setUp(t)
	sqsAPI.Messages["msg1"] = &sqs.Message{Body: aws.String(`{"Message":"{\"Services\":[\"test1\"]}"}`)}

	sub := r.Subscribe()

	select {
	case se := <-sub:
		assert.Equal(t, "test1", se.Name)
	case <-time.After(1000 * time.Millisecond):
		t.Fail()
	}
}
