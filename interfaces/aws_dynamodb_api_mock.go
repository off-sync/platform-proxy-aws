package interfaces

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// AwsDynamoDBAPIMock provides a mock for the AwsDynamoDBAPI to facilitate
// testing.
type AwsDynamoDBAPIMock struct {
	FailDescribeTable bool
	FailScanAllItems  bool
	FailGetItem       bool

	Tables map[string][]map[string]*dynamodb.AttributeValue
}

// NewAwsDynamoDBAPIMock creates a new AwsDynamoDBAPIMock with the Tables
// member initialized.
func NewAwsDynamoDBAPIMock() *AwsDynamoDBAPIMock {
	return &AwsDynamoDBAPIMock{
		Tables: make(map[string][]map[string]*dynamodb.AttributeValue),
	}
}

// SetTable sets the items for the provided table name.
func (m *AwsDynamoDBAPIMock) SetTable(tableName string, items ...map[string]*dynamodb.AttributeValue) {
	m.Tables[tableName] = make([]map[string]*dynamodb.AttributeValue, len(items))

	for i, item := range items {
		m.Tables[tableName][i] = item
	}
}

// DescribeTable returns the description for a table.
func (m *AwsDynamoDBAPIMock) DescribeTable(tableName string) (*dynamodb.TableDescription, error) {
	if m.FailDescribeTable {
		return nil, fmt.Errorf("DescribeTable(%s)", tableName)
	}

	_, found := m.Tables[tableName]
	if !found {
		return nil, fmt.Errorf("table not found: %s", tableName)
	}

	return &dynamodb.TableDescription{
		TableName: aws.String(tableName),
	}, nil
}

// ScanAllItems returns the specified attributes for all items in the provided
// table.
func (m *AwsDynamoDBAPIMock) ScanAllItems(tableName string, attributeNames ...string) ([]map[string]*dynamodb.AttributeValue, error) {
	if m.FailScanAllItems {
		return nil, fmt.Errorf("ScanAllItems(%s, %v)", tableName, attributeNames)
	}

	items, found := m.Tables[tableName]
	if !found {
		return nil, fmt.Errorf("table not found: %s", tableName)
	}

	var f func(map[string]*dynamodb.AttributeValue, []string) map[string]*dynamodb.AttributeValue
	if len(attributeNames) < 1 {
		f = all
	} else {
		f = filter
	}

	for i, item := range items {
		items[i] = f(item, attributeNames)
	}

	return items, nil
}

func all(item map[string]*dynamodb.AttributeValue, attributeNames []string) map[string]*dynamodb.AttributeValue {
	return item
}

func filter(item map[string]*dynamodb.AttributeValue, attributeNames []string) map[string]*dynamodb.AttributeValue {
	out := make(map[string]*dynamodb.AttributeValue)

	for _, name := range attributeNames {
		if v, found := item[name]; found {
			out[name] = v
		}
	}

	return out
}

// GetItem returns a single item from the provided table using the key fields.
func (m *AwsDynamoDBAPIMock) GetItem(tableName string, key map[string]*dynamodb.AttributeValue) (map[string]*dynamodb.AttributeValue, error) {
	if m.FailGetItem {
		return nil, fmt.Errorf("GetItem(%s, %v)", tableName, key)
	}

	items, err := m.ScanAllItems(tableName)
	if err != nil {
		return nil, err
	}

rangeItems:
	for _, item := range items {
		for k, v := range key {
			vv, found := item[k]
			if !found {
				continue rangeItems
			}

			if v.GoString() != vv.GoString() {
				continue rangeItems
			}
		}

		return item, nil
	}

	return nil, nil
}
