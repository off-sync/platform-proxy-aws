package interfaces

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// AwsSqsAPI defines an interface for interacting with the AWS SQS service.
type AwsSqsAPI interface {
	//GetQueueURL returns the URL for the queue with the provided name.
	GetQueueURL(queueName string) (string, error)

	// ReceiveMessageWithContext returns the available message on the queue. It
	// accepts a Context as its first parameter to allow cancellation of the
	// request.
	//
	// It must return an empty slice if no messages could be received before the
	// context was cancelled.
	ReceiveMessageWithContext(ctx context.Context, queueURL string) ([]*sqs.Message, error)

	// DeleteMessage removes the message associated with the provided receipt
	// handle.
	DeleteMessage(queueURL, receiptHandle string) error
}
