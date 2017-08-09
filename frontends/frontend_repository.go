package frontends

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	appInterfaces "github.com/off-sync/platform-proxy-app/interfaces"
	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/off-sync/platform-proxy-domain/frontends"
)

// DynamoDB attribute names.
const (
	AttributeNameFrontendName = "Name"
)

// FrontendRepository implements the FrontendRepository interface backed by an
// AWS DynamoDB table.
type FrontendRepository struct {
	api       interfaces.AwsDynamoDBAPI
	tableName string
}

// NewFrontendRepository creates a new FrontendRepository using the provided AWS DyanamoDB
// API and table name.
func NewFrontendRepository(api interfaces.AwsDynamoDBAPI, tableName string) (*FrontendRepository, error) {
	_, err := api.DescribeTable(tableName)
	if err != nil {
		return nil, err
	}

	return &FrontendRepository{
		api:       api,
		tableName: tableName,
	}, nil
}

// ListFrontends returns all frontend names contained in this repository.
func (r *FrontendRepository) ListFrontends() ([]string, error) {
	items, err := r.api.ScanAllItems(r.tableName, AttributeNameFrontendName)
	if err != nil {
		return nil, err
	}

	var names []string

	var s string
	for _, item := range items {
		name, found := item[AttributeNameFrontendName]
		if !found {
			// skip items without a frontend name
			continue
		}

		err = dynamodbattribute.Unmarshal(name, &s)
		if err != nil {
			// skip items with a non-string frontend name
			continue
		}

		names = append(names, s)
	}

	return names, nil
}

type frontendItem struct {
	Name                 string                     `dynamodbav:"Name"`
	URL                  string                     `dynamodbav:"URL"`
	Certificate          string                     `dynamodbav:"Certificate"`
	PrivateKey           string                     `dynamodbav:"PrivateKey"`
	CertificateExpiresAt dynamodbattribute.UnixTime `dynamodbav:"CertificateExpiresAt"`
	ServiceName          string                     `dynamodbav:"ServiceName"`
}

// DescribeFrontend returns the frontend with the specified name. If no
// frontend	exists with that name an ErrUnknownFrontend is returned.
func (r *FrontendRepository) DescribeFrontend(name string) (*frontends.Frontend, error) {
	keyMap := make(map[string]interface{})
	keyMap[AttributeNameFrontendName] = name

	key, _ := dynamodbattribute.MarshalMap(keyMap) // should not fail

	item, err := r.api.GetItem(r.tableName, key)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, appInterfaces.ErrUnknownFrontend
	}

	f := &frontendItem{}
	dynamodbattribute.UnmarshalMap(item, f) // should not fail

	url, err := url.Parse(f.URL)
	if err != nil {
		return nil, err
	}

	var cert *frontends.Certificate
	if len(f.Certificate) > 0 && len(f.PrivateKey) > 0 {
		cert, err = frontends.NewCertificate([]byte(f.Certificate), []byte(f.PrivateKey))
		if err != nil {
			return nil, err
		}
	}

	return &frontends.Frontend{
		Name:        f.Name,
		URL:         url,
		ServiceName: f.ServiceName,
		Certificate: cert,
	}, nil
}
