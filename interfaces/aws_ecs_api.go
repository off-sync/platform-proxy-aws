package interfaces

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/ecs"
)

// Errors.
var (
	ErrServiceNotFound        = errors.New("service not found")
	ErrTaskDefinitionNotFound = errors.New("task definition not found")
)

// AwsEcsAPI abstracts the use of the AWS ECS API.
type AwsEcsAPI interface {
	// ListServices returns the service arns of the current cluster.
	ListServices() ([]string, error)

	// DescribeService returns the service description for a single service.
	// Returns ErrServiceNotFound if the service is not found.
	DescribeService(serviceArn string) (*ecs.Service, error)

	// DescribeTaskDefinition returns the task definition for the provided arn.
	// Returns ErrTaskDefinitionNotFound if the task definition is not found.
	DescribeTaskDefinition(taskDefArn string) (*ecs.TaskDefinition, error)
}
