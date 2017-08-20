package infra

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
)

var (
	awsSessionOnce sync.Once
	awsSession     *session.Session
	awsSessionErr  error
)

func getSession() (*session.Session, error) {
	awsSessionOnce.Do(func() {
		awsSession, awsSessionErr = session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(viper.GetString(awsID), viper.GetString(awsSecret), ""),
			Region:      aws.String(viper.GetString(awsRegion)),
		})
	})

	return awsSession, awsSessionErr
}
