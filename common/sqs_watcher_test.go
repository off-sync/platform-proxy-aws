package common

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/stretchr/testify/assert"
)

func setUpSqsWatcher(ctx context.Context, t *testing.T) (*SqsWatcher, *interfaces.AwsSqsAPIMock) {
	api := interfaces.NewAwsSqsAPIMock()

	sw, err := NewSqsWatcher(ctx, api, "name")
	assert.Nil(t, err)
	assert.NotNil(t, sw)

	return sw, api
}

func TestNewSqsWatcher(t *testing.T) {
	setUpSqsWatcher(context.Background(), t)
}

func TestNewSqsWatcherShouldAcceptNilContext(t *testing.T) {
	_, err := NewSqsWatcher(nil, interfaces.NewAwsSqsAPIMock(), "name")
	assert.Nil(t, err)
}

func TestNewSqsWatcherShouldReturnErrorOnAPIMissing(t *testing.T) {
	_, err := NewSqsWatcher(context.Background(), nil, "name")
	assert.Equal(t, ErrAwsSqsAPIMissing, err)
}

func TestNewSqsWatcherShouldReturnErrorOnMissingQueueName(t *testing.T) {
	_, err := NewSqsWatcher(context.Background(), interfaces.NewAwsSqsAPIMock(), "")
	assert.Equal(t, ErrAwsSqsQueueNameMissing, err)
}

func TestNewSqsWatcherShouldReturnErrorOnFailingGetQueueURL(t *testing.T) {
	sqsAPI := interfaces.NewAwsSqsAPIMock()
	sqsAPI.FailGetQueueURL = true

	_, err := NewSqsWatcher(context.Background(), sqsAPI, "name")
	assert.NotNil(t, err)
}

func TestCancelSqsWatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	setUpSqsWatcher(ctx, t)

	cancel()
}

func TestReceiveMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sw, mock := setUpSqsWatcher(ctx, t)
	mock.Messages["test1"] = &sqs.Message{Body: aws.String("msg1")}

	sub := sw.Subscribe()

	select {
	case e := <-sub:
		msg, ok := e.(*sqs.Message)
		assert.True(t, ok)
		assert.Equal(t, "msg1", aws.StringValue(msg.Body))
	}

	cancel()
}
