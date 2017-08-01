package interfaces

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// AwsDynamoDBAPI abstracts the use of the AWS DynamoDB API.
type AwsDynamoDBAPI interface {
	// DescribeTable returns the description for a table.
	DescribeTable(tableName string) (*dynamodb.TableDescription, error)

	// ScanAllItems returns the specified attributes for all items in the provided
	// table.
	ScanAllItems(tableName string, attributeNames ...string) ([]map[string]*dynamodb.AttributeValue, error)

	// GetItem returns a single item from the provided table using the key fields.
	GetItem(tableName string, key map[string]*dynamodb.AttributeValue) (map[string]*dynamodb.AttributeValue, error)
}
