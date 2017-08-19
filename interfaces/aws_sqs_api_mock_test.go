package interfaces

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

func setUpAwsSqsAPIMock(t *testing.T) *AwsSqsAPIMock {
	m := NewAwsSqsAPIMock()
	assert.NotNil(t, m)

	return m
}

func TestFailReceiveMessageWithContext(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)
	m.FailReceiveMessageWithContext = true

	_, err := m.ReceiveMessageWithContext(context.Background())
	assert.NotNil(t, err)
}

func TestReceiveMessageWithContext(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)

	m.Messages["test1"] = &sqs.Message{Body: aws.String("msg1")}
	m.Messages["test2"] = &sqs.Message{Body: aws.String("msg2")}
	m.Messages["test3"] = &sqs.Message{Body: aws.String("msg3")}

	msgs, err := m.ReceiveMessageWithContext(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(msgs))

	assert.Equal(t, "msg1", aws.StringValue(msgs[0].Body))
	assert.Equal(t, "msg2", aws.StringValue(msgs[1].Body))
	assert.Equal(t, "msg3", aws.StringValue(msgs[2].Body))
}

func TestFailDeleteMessage(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)
	m.FailDeleteMessage = true

	err := m.DeleteMessage("")
	assert.NotNil(t, err)
}

func TestDeleteMessageShouldFailOnUnknownReceiptHandle(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)

	err := m.DeleteMessage("test")
	assert.NotNil(t, err)
}

func TestDeleteMessage(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)
	m.Messages["test"] = &sqs.Message{}

	err := m.DeleteMessage("test")
	assert.Nil(t, err)
}
