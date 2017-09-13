package infra

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/off-sync/platform-proxy-aws/interfaces"
)

// AwsDynamoDBSdk implements the AWS DynamoDB API.
type AwsDynamoDBSdk struct {
	dynSvc *dynamodb.DynamoDB
}

// NewAwsDynamoDBSdk creates a new AWS DynamoDB SDK.
func NewAwsDynamoDBSdk(dynSvc *dynamodb.DynamoDB) *AwsDynamoDBSdk {
	return &AwsDynamoDBSdk{
		dynSvc: dynSvc,
	}
}

// NewAwsDynamoDBSdkFromConfig creates a new AwsDynamoDBSdk using the configuration
// exposed via viper. The AWS ID, secret, region are retrieved from the
// configuration.
func NewAwsDynamoDBSdkFromConfig(config interfaces.Config) (*AwsDynamoDBSdk, error) {
	sess, err := getSession(config)
	if err != nil {
		return nil, err
	}

	dynSvc := dynamodb.New(sess)

	return NewAwsDynamoDBSdk(dynSvc), nil
}

// DescribeTable returns the description for a table.
func (s *AwsDynamoDBSdk) DescribeTable(tableName string) (*dynamodb.TableDescription, error) {
	dt, err := s.dynSvc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	return dt.Table, nil
}

// ScanAllItems returns the specified attributes for all items in the provided
// table.
func (s *AwsDynamoDBSdk) ScanAllItems(tableName string, attributeNames ...string) ([]map[string]*dynamodb.AttributeValue, error) {
	var items []map[string]*dynamodb.AttributeValue

	err := s.dynSvc.ScanPages(&dynamodb.ScanInput{
		TableName:       aws.String(tableName),
		AttributesToGet: aws.StringSlice(attributeNames),
	}, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		for _, item := range page.Items {
			items = append(items, item)
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetItem returns a single item from the provided table using the key fields.
func (s *AwsDynamoDBSdk) GetItem(tableName string, key map[string]*dynamodb.AttributeValue) (map[string]*dynamodb.AttributeValue, error) {
	i, err := s.dynSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}

	return i.Item, nil
}
