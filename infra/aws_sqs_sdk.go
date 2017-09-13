package infra

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/off-sync/platform-proxy-aws/interfaces"
)

// Errors.
var (
	ErrQueueNotFound       = errors.New("Queue not found")
	ErrMultipleQueuesFound = errors.New("Multiple queues found, expected 1")
)

// Defaults.
const (
	DefaultWaitTime = 10
)

// AwsSqsSdk implements the AWS SQS API.
type AwsSqsSdk struct {
	sqsSvc *sqs.SQS

	// configuration
	waitTime          int
	visibilityTimeout int
}

// AwsSqsSdkOption defines the interface for options that can configure
// this SDK.
type AwsSqsSdkOption func(*AwsSqsSdk) error

// NewAwsSqsSdk creates a new AwsSqsSdk.
func NewAwsSqsSdk(sqsSvc *sqs.SQS, options ...AwsSqsSdkOption) (*AwsSqsSdk, error) {
	s := &AwsSqsSdk{
		sqsSvc:   sqsSvc,
		waitTime: DefaultWaitTime,
	}

	for _, opt := range options {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// NewAwsSqsSdkFromConfig creates a new AWS SQS SDK from configuration.
func NewAwsSqsSdkFromConfig(config interfaces.Config) (*AwsSqsSdk, error) {
	sess, err := getSession(config)
	if err != nil {
		return nil, err
	}

	sqsSvc := sqs.New(sess)

	return NewAwsSqsSdk(sqsSvc)
}

// WithWaitTime configures the AwsSqsSdk with the provided wait time in seconds.
func WithWaitTime(seconds int) AwsSqsSdkOption {
	return func(s *AwsSqsSdk) error {
		s.waitTime = seconds

		// set visibility timeout to half the wait time, with a minimum of 1 second
		s.visibilityTimeout = seconds / 2
		if s.visibilityTimeout < 1 {
			s.visibilityTimeout = 1
		}

		return nil
	}
}

//GetQueueURL returns the URL for the queue with the provided name.
func (s *AwsSqsSdk) GetQueueURL(queueName string) (string, error) {
	lqo, err := s.sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(queueName),
	})
	if err != nil {
		return "", err
	}

	if len(lqo.QueueUrls) < 1 {
		return "", ErrQueueNotFound
	} else if len(lqo.QueueUrls) > 1 {
		return "", ErrMultipleQueuesFound
	}

	return aws.StringValue(lqo.QueueUrls[0]), nil
}

// ReceiveMessageWithContext returns the available messages on the queue. It
// accepts a Context as its first parameter to allow cancellation of the
// request.
//
// It must return an empty slice if no messages could be received before the
// context was cancelled.
func (s *AwsSqsSdk) ReceiveMessageWithContext(ctx context.Context, queueURL string) ([]*sqs.Message, error) {
	rmo, err := s.sqsSvc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:          aws.String(queueURL),
		VisibilityTimeout: aws.Int64(int64(s.visibilityTimeout)),
		WaitTimeSeconds:   aws.Int64(int64(s.waitTime)),
	})
	if err != nil {
		return nil, err
	}

	return rmo.Messages, nil
}

// DeleteMessage removes a message from the provided queue.
func (s *AwsSqsSdk) DeleteMessage(queueURL, receiptHandle string) error {
	_, err := s.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})

	return err
}
