package infra

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func initConfig(t *testing.T) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".platform-proxy-aws")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		t.Logf("using config file: %s", viper.ConfigFileUsed())
	}
}

func setUp(t *testing.T) *ServiceRepository {
	initConfig(t)

	awsRegion := viper.GetString("awsRegion")
	assert.NotEmpty(t, awsRegion)

	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	assert.Nil(t, err)

	ecsSvc := ecs.New(sess)
	assert.NotNil(t, ecsSvc)

	clusterName := viper.GetString("ecsClusterName")
	assert.NotEmpty(t, clusterName)

	serviceRepository, err := NewServiceRepository(
		ecsSvc,
		clusterName,
		WithServerContainerName("server"),
		WithDockerLabelPort("com.off-sync.platform.proxy.port"),
		WithDefaultPort(8080))

	assert.Nil(t, err)
	assert.NotNil(t, serviceRepository)

	return serviceRepository
}

func TestNewServiceRepository(t *testing.T) {
	setUp(t)
}

func TestListServices(t *testing.T) {
	r := setUp(t)

	names, err := r.ListServices()
	assert.Nil(t, err)

	for _, name := range names {
		t.Log(name)
	}
}

func TestDescribeService(t *testing.T) {
	r := setUp(t)

	names, err := r.ListServices()
	assert.Nil(t, err)

	for _, name := range names {
		service, err := r.DescribeService(name)
		assert.Nil(t, err)
		assert.NotNil(t, service)

		t.Logf("%s: %v", service.Name, service.Servers)
	}
}
