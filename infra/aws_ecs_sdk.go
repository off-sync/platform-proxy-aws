package infra

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy-aws/interfaces"
)

// AwsEcsSdk implements the AwsEcsAPI.
type AwsEcsSdk struct {
	ecsSvc  *ecs.ECS
	cluster *ecs.Cluster
}

// NewAwsEcsSdk creates a new AwsEcsSdk using the provided ECS service and
// cluster name. It returns an error if the cluster cannot be described.
func NewAwsEcsSdk(ecsSvc *ecs.ECS, clusterName string) (*AwsEcsSdk, error) {
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

	return &AwsEcsSdk{
		ecsSvc:  ecsSvc,
		cluster: clusters.Clusters[0],
	}, nil
}

// ListServices returns the service arns of the current cluster.
func (s *AwsEcsSdk) ListServices() ([]string, error) {
	var serviceNames []string

	err := s.ecsSvc.ListServicesPages(&ecs.ListServicesInput{
		Cluster: s.cluster.ClusterName,
	}, func(output *ecs.ListServicesOutput, lastPage bool) bool {
		serviceNames = append(serviceNames, aws.StringValueSlice(output.ServiceArns)...)
		return true
	})
	if err != nil {
		return nil, err
	}

	// remove ARN prefix
	for i, name := range serviceNames {
		serviceNames[i] = name[strings.LastIndex(name, "/")+1:]
	}

	return serviceNames, nil
}

// DescribeService returns the service description for a single service.
func (s *AwsEcsSdk) DescribeService(serviceArn string) (*ecs.Service, error) {
	serviceDescription, err := s.ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  s.cluster.ClusterName,
		Services: aws.StringSlice([]string{serviceArn}),
	})
	if err != nil {
		return nil, err
	}

	if len(serviceDescription.Services) < 1 {
		return nil, interfaces.ErrServiceNotFound
	}

	return serviceDescription.Services[0], nil
}

// DescribeTaskDefinition returns the task definition for the provided arn.
func (s *AwsEcsSdk) DescribeTaskDefinition(taskDefArn string) (*ecs.TaskDefinition, error) {
	tdef, err := s.ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefArn),
	})
	if err != nil {
		return nil, err
	}

	return tdef.TaskDefinition, nil
}

// NewAwsEcsSdkFromConfig creates a new AwsEcsSdk using the configuration
// exposed via viper. The AWS ID, secret, region and cluster name are retrieved
// from the configuration.
func NewAwsEcsSdkFromConfig(config interfaces.Config) (*AwsEcsSdk, error) {
	sess, err := getSession(config)
	if err != nil {
		return nil, err
	}

	ecsSvc := ecs.New(sess)

	return NewAwsEcsSdk(ecsSvc, config.GetString(ecsClusterName))
}
