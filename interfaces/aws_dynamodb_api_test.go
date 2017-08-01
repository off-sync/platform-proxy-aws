package interfaces

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/stretchr/testify/assert"
)

func setUp(t *testing.T) *AwsDynamoDBAPIMock {
	return NewAwsDynamoDBAPIMock()
}

func TestNewAwsDynamoDBAPIMock(t *testing.T) {
	m := setUp(t)
	assert.NotNil(t, m)
}

func TestAwsDynamoDBAPIMockSetTable(t *testing.T) {
	m := setUp(t)

	item1 := make(map[string]*dynamodb.AttributeValue)
	item2 := make(map[string]*dynamodb.AttributeValue)
	item3 := make(map[string]*dynamodb.AttributeValue)

	m.SetTable("table1", item1, item2, item3)

	assert.EqualValues(t, []map[string]*dynamodb.AttributeValue{item1, item2, item3}, m.Tables["table1"])
}

func TestAwsDynamoDBAPIMockDescribeTable(t *testing.T) {
	m := setUp(t)

	m.SetTable("table1")

	td, err := m.DescribeTable("table1")
	assert.Nil(t, err)
	assert.Equal(t, "table1", aws.StringValue(td.TableName))
}

func TestAwsDynamoDBAPIMockDescribeTableFails(t *testing.T) {
	m := setUp(t)

	m.SetTable("table1")

	m.FailDescribeTable = true

	_, err := m.DescribeTable("table1")
	assert.NotNil(t, err)
}

func TestAwsDynamoDBAPIMockDescribeTableShouldReturnErrorForNonExistingTable(t *testing.T) {
	m := setUp(t)

	td, err := m.DescribeTable("table1")
	assert.NotNil(t, err)
	assert.Nil(t, td)
}

func TestAwsDynamoDBAPIMockScanAllItems(t *testing.T) {
	m := setUp(t)

	item1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "1-value1",
		"attr2": "1-value2",
	})
	assert.Nil(t, err)

	item2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
		"attr2": "2-value2",
	})
	assert.Nil(t, err)

	item3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
		"attr2": "3-value2",
	})
	assert.Nil(t, err)

	m.SetTable("table1", item1, item2, item3)

	items, err := m.ScanAllItems("table1")
	assert.Nil(t, err)

	assert.EqualValues(t, []map[string]*dynamodb.AttributeValue{item1, item2, item3}, items)
}

func TestAwsDynamoDBAPIMockScanAllItemsFails(t *testing.T) {
	m := setUp(t)

	m.SetTable("table1")

	m.FailScanAllItems = true

	_, err := m.ScanAllItems("table1")
	assert.NotNil(t, err)
}

func TestAwsDynamoDBAPIMockScanAllItemsReturnsErrorForNonExistingTable(t *testing.T) {
	m := setUp(t)

	items, err := m.ScanAllItems("table1")
	assert.NotNil(t, err)
	assert.Nil(t, items)
}

func TestAwsDynamoDBAPIMockScanAllItemsOnlySpecifiedAttributes(t *testing.T) {
	m := setUp(t)

	item1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "1-value1",
		"attr2": "1-value2",
	})
	assert.Nil(t, err)

	item2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
		"attr2": "2-value2",
	})
	assert.Nil(t, err)

	item3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
		"attr2": "3-value2",
	})
	assert.Nil(t, err)

	m.SetTable("table1", item1, item2, item3)

	items, err := m.ScanAllItems("table1", "attr1")
	assert.Nil(t, err)

	exp1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "1-value1",
	})
	assert.Nil(t, err)

	exp2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
	})
	assert.Nil(t, err)

	exp3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
	})
	assert.Nil(t, err)

	assert.EqualValues(t, []map[string]*dynamodb.AttributeValue{exp1, exp2, exp3}, items)
}

func TestAwsDynamoDBAPIMockGetItem(t *testing.T) {
	m := setUp(t)

	item1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr2": "1-value2",
	})
	assert.Nil(t, err)

	item2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
		"attr2": "2-value2",
	})
	assert.Nil(t, err)

	item3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
		"attr2": "3-value2",
	})
	assert.Nil(t, err)

	m.SetTable("table1", item1, item2, item3)

	key, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
	})
	assert.Nil(t, err)

	item, err := m.GetItem("table1", key)
	assert.Nil(t, err)

	assert.EqualValues(t, item3, item)
}

func TestAwsDynamoDBAPIMockGetItemFails(t *testing.T) {
	m := setUp(t)

	item1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "1-value1",
		"attr2": "1-value2",
	})
	assert.Nil(t, err)

	item2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
		"attr2": "2-value2",
	})
	assert.Nil(t, err)

	item3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "3-value1",
		"attr2": "3-value2",
	})
	assert.Nil(t, err)

	m.SetTable("table1", item1, item2, item3)

	key, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
	})
	assert.Nil(t, err)

	m.FailGetItem = true

	_, err = m.GetItem("table1", key)
	assert.NotNil(t, err)

	m.FailGetItem = false
	m.FailScanAllItems = true

	_, err = m.GetItem("table1", key)
	assert.NotNil(t, err)
}

func TestAwsDynamoDBAPIMockGetItemReturnsNilIfItemNotFound(t *testing.T) {
	m := setUp(t)

	m.SetTable("table1")

	key, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"attr1": "2-value1",
	})
	assert.Nil(t, err)

	item, err := m.GetItem("table1", key)
	assert.Nil(t, err)

	assert.Nil(t, item)
}
