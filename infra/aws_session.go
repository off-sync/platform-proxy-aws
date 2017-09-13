package infra

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/off-sync/platform-proxy-aws/interfaces"
)

var (
	awsSessionOnce sync.Once
	awsSession     *session.Session
	awsSessionErr  error
)

func getSession(config interfaces.Config) (*session.Session, error) {
	awsSessionOnce.Do(func() {
		cfg := &aws.Config{}

		if (config.GetString(awsID) != "") &&
			(config.GetString(awsSecret) != "") {
			cfg.Credentials = credentials.NewStaticCredentials(config.GetString(awsID), config.GetString(awsSecret), "")
		}

		if config.GetString(awsRegion) != "" {
			cfg.Region = aws.String(config.GetString(awsRegion))
		}

		awsSession, awsSessionErr = session.NewSession(cfg)
	})

	return awsSession, awsSessionErr
}
