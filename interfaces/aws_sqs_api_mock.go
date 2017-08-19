package interfaces

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// AwsSqsAPIMock mocks the AWS SQS API by providing flags that determine
// whether methods should always fail, and exposing the various return values
// in public members of the struct.
type AwsSqsAPIMock struct {
	// Flags that determine whether an error will always be returned.
	FailReceiveMessageWithContext bool
	FailDeleteMessage             bool

	// Map of receipt handles to messages
	Messages map[string]*sqs.Message
}

// NewAwsSqsAPIMock creates a new AWS SQS API mock.
func NewAwsSqsAPIMock() *AwsSqsAPIMock {
	return &AwsSqsAPIMock{
		Messages: make(map[string]*sqs.Message),
	}
}

// ReceiveMessageWithContext returns the available message on the queue. It
// accepts a Context as its first parameter to allow cancellation of the
// request.
//
// It must return an empty slice if no messages could be received before the
// context was cancelled.
func (m *AwsSqsAPIMock) ReceiveMessageWithContext(ctx context.Context) ([]*sqs.Message, error) {
	if m.FailReceiveMessageWithContext {
		return nil, fmt.Errorf("ReceiveMessageWithContext(%v): fail", ctx)
	}

	var msgs []*sqs.Message
	for _, msg := range m.Messages {
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// DeleteMessage removes the message associated with the provided receipt
// handle.
func (m *AwsSqsAPIMock) DeleteMessage(receiptHandle string) error {
	if m.FailDeleteMessage {
		return fmt.Errorf("DeleteMessage(%s): fail", receiptHandle)
	}

	if _, found := m.Messages[receiptHandle]; !found {
		return fmt.Errorf("DeleteMessage(%s): unknown receipt handle", receiptHandle)
	}

	return nil
}
