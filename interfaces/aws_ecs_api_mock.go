package interfaces

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ecs"
)

// AwsEcsAPIMock mocks the AWS ECS API by providing flags that determine
// whether method calls always fail, and exposing the various return values
// in public members of the struct.
type AwsEcsAPIMock struct {
	// Flags that determine whether an error will always be returned.
	FailListServices           bool
	FailDescribeService        bool
	FailDescribeTaskDefinition bool

	// Return values.
	ServiceNames []string
	Services     map[string]*ecs.Service
	TaskDefs     map[string]*ecs.TaskDefinition
}

// NewAwsEcsAPIMock creates a new AWS ECS API mock with initialized map
// members.
func NewAwsEcsAPIMock() *AwsEcsAPIMock {
	return &AwsEcsAPIMock{
		Services: make(map[string]*ecs.Service),
		TaskDefs: make(map[string]*ecs.TaskDefinition),
	}
}

// ListServices returns the service arns of the current cluster.
func (m *AwsEcsAPIMock) ListServices() ([]string, error) {
	if m.FailListServices {
		return nil, fmt.Errorf("%+v.ListServices()", m)
	}

	return m.ServiceNames, nil
}

// DescribeService returns the service description for a single service.
func (m *AwsEcsAPIMock) DescribeService(serviceArn string) (*ecs.Service, error) {
	if m.FailDescribeService {
		return nil, fmt.Errorf("%+v.DescribeService(%s)", m, serviceArn)
	}

	s, found := m.Services[serviceArn]
	if !found {
		return nil, ErrServiceNotFound
	}

	return s, nil
}

// DescribeTaskDefinition returns the task definition for the provided arn.
func (m *AwsEcsAPIMock) DescribeTaskDefinition(taskDefArn string) (*ecs.TaskDefinition, error) {
	if m.FailDescribeTaskDefinition {
		return nil, fmt.Errorf("%+v.DescribeTaskDefinition(%s)", m, taskDefArn)
	}

	s, found := m.TaskDefs[taskDefArn]
	if !found {
		return nil, ErrTaskDefinitionNotFound
	}

	return s, nil
}
