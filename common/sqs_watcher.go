package common

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/off-sync/platform-proxy-aws/interfaces"
)

// Errors.
var (
	ErrAwsSqsAPIMissing       = errors.New("AWS SQS API missing")
	ErrAwsSqsQueueNameMissing = errors.New("AWS SQS Queue Name missing")
)

// SqsWatcher uses the provided AWS SQS API to listen for messages and forwards
// the received messages to the registered subscriptions.
type SqsWatcher struct {
	*SubscriptionManager

	api      interfaces.AwsSqsAPI
	queueURL string
}

// NewSqsWatcher creates a new SQS Watcher using the provided context for
// cancellation, and API to query the AWS SQS.
func NewSqsWatcher(
	ctx context.Context,
	api interfaces.AwsSqsAPI,
	queueName string) (*SqsWatcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if api == nil {
		return nil, ErrAwsSqsAPIMissing
	}

	if queueName == "" {
		return nil, ErrAwsSqsQueueNameMissing
	}

	queueURL, err := api.GetQueueURL(queueName)
	if err != nil {
		return nil, err
	}

	sw := &SqsWatcher{
		SubscriptionManager: NewSubscriptionManager(ctx),
		api:                 api,
		queueURL:            queueURL,
	}

	// start message polling
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if msgs, err := sw.api.ReceiveMessageWithContext(ctx, sw.queueURL); err == nil {
					for _, msg := range msgs {
						sw.Publish(msg)

						sw.api.DeleteMessage(sw.queueURL, aws.StringValue(msg.ReceiptHandle))
					}
				}
			}
		}
	}()

	return sw, nil
}
