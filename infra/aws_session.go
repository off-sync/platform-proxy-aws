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
		cfg := &aws.Config{}

		if viper.GetString(awsID) != "" {
			cfg.Credentials = credentials.NewStaticCredentials(viper.GetString(awsID), viper.GetString(awsSecret), "")
		}

		if viper.GetString(awsRegion) != "" {
			cfg.Region = aws.String(viper.GetString(awsRegion))
		}

		awsSession, awsSessionErr = session.NewSession(cfg)
	})

	return awsSession, awsSessionErr
}
