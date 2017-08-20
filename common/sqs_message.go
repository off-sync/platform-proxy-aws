package common

// SqsMessageBody represents the Body structure of an SQS Message.
type SqsMessageBody struct {
	Message string `json:"Message"`
}
