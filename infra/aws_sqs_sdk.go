package infra

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
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
	sqsSvc   *sqs.SQS
	queueURL *string

	// configuration
	waitTime          int
	visibilityTimeout int
}

// AwsSqsSdkOption defines the interface for options that can configure
// this SDK.
type AwsSqsSdkOption func(*AwsSqsSdk) error

// NewAwsSqsSdk creates a new AwsSqsSdk.
func NewAwsSqsSdk(sqsSvc *sqs.SQS, queueName string, options ...AwsSqsSdkOption) (*AwsSqsSdk, error) {
	lqo, err := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(queueName),
	})
	if err != nil {
		return nil, err
	}

	if len(lqo.QueueUrls) < 1 {
		return nil, ErrQueueNotFound
	} else if len(lqo.QueueUrls) > 1 {
		return nil, ErrMultipleQueuesFound
	}

	s := &AwsSqsSdk{
		sqsSvc:   sqsSvc,
		queueURL: lqo.QueueUrls[0],
		waitTime: DefaultWaitTime,
	}

	for _, opt := range options {
		err = opt(s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
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

// ReceiveMessageWithContext returns the available messages on the queue. It
// accepts a Context as its first parameter to allow cancellation of the
// request.
//
// It must return an empty slice if no messages could be received before the
// context was cancelled.
func (s *AwsSqsSdk) ReceiveMessageWithContext(ctx context.Context) ([]*sqs.Message, error) {
	rmo, err := s.sqsSvc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:          s.queueURL,
		VisibilityTimeout: aws.Int64(int64(s.visibilityTimeout)),
		WaitTimeSeconds:   aws.Int64(int64(s.waitTime)),
	})
	if err != nil {
		return nil, err
	}

	return rmo.Messages, nil
}

// DeleteMessage removes a message from the provided queue.
func (s *AwsSqsSdk) DeleteMessage(receiptHandle string) error {
	_, err := s.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      s.queueURL,
		ReceiptHandle: aws.String(receiptHandle),
	})

	return err
}
