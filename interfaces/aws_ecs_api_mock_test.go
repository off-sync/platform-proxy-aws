package interfaces

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

func TestNewAwsEcsAPIMock(t *testing.T) {
	m := NewAwsEcsAPIMock()
	assert.NotNil(t, m)
}

func TestAwsEcsAPIMockFails(t *testing.T) {
	m := NewAwsEcsAPIMock()
	m.FailListServices = true
	_, err := m.ListServices()
	assert.NotNil(t, err)

	m = NewAwsEcsAPIMock()
	m.FailDescribeService = true
	_, err = m.DescribeService("serviceArn")
	assert.NotNil(t, err)

	m = NewAwsEcsAPIMock()
	m.FailDescribeTaskDefinition = true
	_, err = m.DescribeTaskDefinition("taskDefArn")
	assert.NotNil(t, err)
}

func TestAwsEcsAPIMockReturnsCorrectErrorOnNotFound(t *testing.T) {
	m := NewAwsEcsAPIMock()

	_, err := m.DescribeService("serviceArn")
	assert.Equal(t, ErrServiceNotFound, err)

	_, err = m.DescribeTaskDefinition("taskDefArn")
	assert.Equal(t, ErrTaskDefinitionNotFound, err)
}

func TestAwsEcsAPIMockReturnsConfiguredReturnValues(t *testing.T) {
	m := NewAwsEcsAPIMock()

	m.ServiceNames = append(m.ServiceNames, "serviceArn")

	svcs, err := m.ListServices()
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"serviceArn"}, svcs)

	expectedSvc := &ecs.Service{}
	m.Services["serviceArn"] = expectedSvc

	svc, err := m.DescribeService("serviceArn")
	assert.Nil(t, err)
	assert.Equal(t, expectedSvc, svc)

	expectedTaskDef := &ecs.TaskDefinition{}
	m.TaskDefs["taskDefArn"] = expectedTaskDef

	taskDef, err := m.DescribeTaskDefinition("taskDefArn")
	assert.Nil(t, err)
	assert.Equal(t, expectedTaskDef, taskDef)
}
