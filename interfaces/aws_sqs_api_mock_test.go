package interfaces

import (
	"context"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestImplementsAPI(t *testing.T) {
	var api interface{}
	api = &AwsSqsAPIMock{}

	_, ok := api.(AwsSqsAPI)
	assert.True(t, ok)
}

func setUpAwsSqsAPIMock(t *testing.T) *AwsSqsAPIMock {
	m := NewAwsSqsAPIMock()
	assert.NotNil(t, m)

	return m
}

func TestGetQueueURLFail(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)
	m.FailGetQueueURL = true

	_, err := m.GetQueueURL("name")
	assert.NotNil(t, err)
}

func TestGetQueueURL(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)

	url, err := m.GetQueueURL("name")
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
}

func TestFailReceiveMessageWithContext(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)
	m.FailReceiveMessageWithContext = true

	_, err := m.ReceiveMessageWithContext(context.Background(), "url")
	assert.NotNil(t, err)
}

func TestReceiveMessageWithContext(t *testing.T) {
	m := setUpAwsSqsAPIMock(t)

	m.Messages["test1"] = &sqs.Message{Body: aws.String("msg1")}
	m.Messages["test2"] = &sqs.Message{Body: aws.String("msg2")}
	m.Messages["test3"] = &sqs.Message{Body: aws.String("msg3")}

	msgs, err := m.ReceiveMessageWithContext(context.Background(), "url")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(msgs))

	// sort received messages on receipt handle
	sort.Sort(byReceiptHandle(msgs))

	assert.Equal(t, "msg1", aws.StringValue(msgs[0].Body))
	assert.Equal(t, "msg2", aws.StringValue(msgs[1].Body))
	assert.Equal(t, "msg3", aws.StringValue(msgs[2].Body))
}

type byReceiptHandle []*sqs.Message

func (s byReceiptHandle) Len() int { return len(s) }
func (s byReceiptHandle) Less(i, j int) bool {
	return aws.StringValue(s[i].ReceiptHandle) < aws.StringValue(s[j].ReceiptHandle)
}
func (s byReceiptHandle) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

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
