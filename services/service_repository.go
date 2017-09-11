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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	appInterfaces "github.com/off-sync/platform-proxy-app/interfaces"
	"github.com/off-sync/platform-proxy-aws/common"
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

	sqsWatcher *common.SqsWatcher
}

// Errors.
var (
	ErrMissingAwsEcsAPI  = errors.New("missing AWS ECS API")
	ErrMissingSqsWatcher = errors.New("missing SQS Watcher")
)

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
func NewServiceRepository(
	api interfaces.AwsEcsAPI,
	sqsWatcher *common.SqsWatcher,
	options ...ServiceRepositoryOption) (*ServiceRepository, error) {
	if api == nil {
		return nil, ErrMissingAwsEcsAPI
	}

	if sqsWatcher == nil {
		return nil, ErrMissingSqsWatcher
	}

	r := &ServiceRepository{
		api:                 api,
		sqsWatcher:          sqsWatcher,
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
		if err == interfaces.ErrServiceNotFound {
			// map API error to interface error
			return nil, appInterfaces.ErrUnknownService
		}

		return nil, err
	}

	if aws.StringValue(service.Status) == "INACTIVE" {
		return nil, appInterfaces.ErrUnknownService
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

// Subscribe returns a channel through which frontend events will
// be distributed.
func (r *ServiceRepository) Subscribe() <-chan appInterfaces.ServiceEvent {
	svcChan := make(chan appInterfaces.ServiceEvent)

	go func() {
		sub := r.sqsWatcher.Subscribe()

		for {
			e := <-sub

			if e == nil {
				// subscription was cancelled
				return
			}

			if sqsMsg, ok := e.(*sqs.Message); ok {
				sendServiceEvent(svcChan, sqsMsg)
			}
		}
	}()

	return svcChan
}

type serviceEventMessage struct {
	Services []string `json:"Services"`
}

func sendServiceEvent(svcChan chan<- appInterfaces.ServiceEvent, sqsMsg *sqs.Message) {
	body := &common.SqsMessageBody{}
	if err := json.Unmarshal([]byte(aws.StringValue(sqsMsg.Body)), body); err != nil {
		return
	}

	msg := &serviceEventMessage{}
	if err := json.Unmarshal([]byte(body.Message), msg); err != nil {
		return
	}

	for _, name := range msg.Services {
		svcChan <- appInterfaces.ServiceEvent{Name: name}
	}
}
