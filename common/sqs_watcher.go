package common

import (
	"context"
	"errors"

	"github.com/off-sync/platform-proxy-aws/interfaces"
)

// Errors.
var (
	ErrAwsSqsAPIMissing = errors.New("AWS SQS API missing")
)

// SqsWatcher uses the provided AWS SQS API to listen for messages and forwards
// the received messages to the registered subscriptions.
type SqsWatcher struct {
	*SubscriptionManager

	api interfaces.AwsSqsAPI
}

// NewSqsWatcher creates a new SQS Watcher using the provided context for
// cancellation, and API to query the AWS SQS.
func NewSqsWatcher(
	ctx context.Context, api interfaces.AwsSqsAPI) (*SqsWatcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if api == nil {
		return nil, ErrAwsSqsAPIMissing
	}

	sw := &SqsWatcher{
		SubscriptionManager: NewSubscriptionManager(ctx),
		api:                 api,
	}

	// start message polling
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if msgs, err := sw.api.ReceiveMessageWithContext(ctx); err == nil {
					for _, msg := range msgs {
						sw.Publish(msg)
					}
				}
			}
		}
	}()

	return sw, nil
}
