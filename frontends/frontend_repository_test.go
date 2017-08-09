package frontends

import (
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/off-sync/platform-proxy-aws/interfaces"
	"github.com/off-sync/platform-proxy-domain/frontends"
	"github.com/stretchr/testify/assert"
)

const tableName = "frontends"

func setUp(t *testing.T) (*FrontendRepository, *interfaces.AwsDynamoDBAPIMock) {
	api := interfaces.NewAwsDynamoDBAPIMock()

	item1, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"Name":        "frontend1",
		"ServiceName": "service1",
		"URL":         "https://frontend1.off-sync.net",
		"Certificate": certPEM,
		"PrivateKey":  keyPEM,
	})
	assert.Nil(t, err)

	item2, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"Name": "frontend2",
		"URL":  "%zzzzz",
	})
	assert.Nil(t, err)

	item3, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"Name":        "frontend3",
		"ServiceName": "service3",
		"URL":         "https://frontend3.off-sync.net",
		"Certificate": "NOT PEM",
		"PrivateKey":  "NOT PEM",
	})
	assert.Nil(t, err)

	item4, err := dynamodbattribute.MarshalMap(map[string]interface{}{})
	assert.Nil(t, err)

	item5, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"Name": true,
	})
	assert.Nil(t, err)

	// frontend without certificate
	item6, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"Name":        "frontend6",
		"ServiceName": "service6",
		"URL":         "https://frontend6.off-sync.net",
	})
	assert.Nil(t, err)

	api.SetTable(tableName, item1, item2, item3, item4, item5, item6)

	r, err := NewFrontendRepository(api, tableName)
	assert.Nil(t, err)

	assert.NotNil(t, r)

	return r, api
}

func TestNewFrontendRepository(t *testing.T) {
	r, _ := setUp(t)
	assert.NotNil(t, r)
}

func TestNewFrontendRepositoryShouldReturnErrorForNonExistingTable(t *testing.T) {
	api := interfaces.NewAwsDynamoDBAPIMock()

	r, err := NewFrontendRepository(api, tableName)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestFrontendRepositoryListFrontends(t *testing.T) {
	r, _ := setUp(t)

	names, err := r.ListFrontends()
	assert.Nil(t, err)

	assert.EqualValues(t, []string{"frontend1", "frontend2", "frontend3", "frontend6"}, names)
}

func TestFrontendRepositoryListFrontendsReturnsErrorWhenScanAllItemsFails(t *testing.T) {
	r, api := setUp(t)

	api.FailScanAllItems = true

	_, err := r.ListFrontends()
	assert.NotNil(t, err)
}

func TestFrontendRepositoryDescribeFrontend(t *testing.T) {
	r, _ := setUp(t)

	f, err := r.DescribeFrontend("frontend1")
	assert.Nil(t, err)

	url, err := url.Parse("https://frontend1.off-sync.net")
	assert.Nil(t, err)

	cert, err := frontends.NewCertificate([]byte(certPEM), []byte(keyPEM))
	assert.Nil(t, err)

	assert.EqualValues(t, &frontends.Frontend{
		Name:        "frontend1",
		ServiceName: "service1",
		URL:         url,
		Certificate: cert,
	}, f)
}

func TestFrontendRepositoryDescribeFrontendWithoutCertificate(t *testing.T) {
	r, _ := setUp(t)

	f, err := r.DescribeFrontend("frontend6")
	assert.Nil(t, err)

	url, err := url.Parse("https://frontend6.off-sync.net")
	assert.Nil(t, err)

	assert.EqualValues(t, &frontends.Frontend{
		Name:        "frontend6",
		ServiceName: "service6",
		URL:         url,
	}, f)
}

func TestFrontendRepositoryDescribeFrontendReturnsErrorForNonExistingFrontend(t *testing.T) {
	r, _ := setUp(t)

	_, err := r.DescribeFrontend("unknown")
	assert.NotNil(t, err)
}

func TestFrontendRepositoryDescribeFrontendReturnsErrorWhenGetItemFails(t *testing.T) {
	r, api := setUp(t)

	api.FailGetItem = true

	_, err := r.DescribeFrontend("frontend1")
	assert.NotNil(t, err)
}

func TestFrontendRepositoryDescribeFrontendReturnsErrorWhenURLIsInvalid(t *testing.T) {
	r, _ := setUp(t)

	_, err := r.DescribeFrontend("frontend2")
	assert.NotNil(t, err)
}

func TestFrontendRepositoryDescribeFrontendReturnsErrorWhenCertificateIsInvalid(t *testing.T) {
	r, _ := setUp(t)

	_, err := r.DescribeFrontend("frontend3")
	assert.NotNil(t, err)
}

// Not After: Jul 20 10:49:01 2027 GMT
//
// Steps used to create the certificate and key:
// openssl genrsa -out ca.key 4096
// openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.pem
// openssl genrsa -out test.key 4096
// openssl req -new -key test.key -out test.csr
// openssl x509 -req -in test.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out test.crt -days 3650 -sha256
// cat ca.pem test.crt > test.pem
var certPEM = `-----BEGIN CERTIFICATE-----
MIIFfTCCA2WgAwIBAgIJAIl+K0g3ga7nMA0GCSqGSIb3DQEBCwUAMFUxCzAJBgNV
BAYTAk5MMRMwEQYDVQQIDApTb21lLVN0YXRlMRUwEwYDVQQKDAxPZmYtU3luYy5j
b20xGjAYBgNVBAMMEXRlc3RAb2ZmLXN5bmMuY29tMB4XDTE3MDcyMjExMjM0MVoX
DTI3MDcyMDExMjM0MVowVTELMAkGA1UEBhMCTkwxEzARBgNVBAgMClNvbWUtU3Rh
dGUxFTATBgNVBAoMDE9mZi1TeW5jLmNvbTEaMBgGA1UEAwwRdGVzdEBvZmYtc3lu
Yy5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDJIXOJIHj+SQt1
sI+xHZuATEen1UsAi5pCzRvJmnrH0CzFXrGjF8Fe6izNxkN0ZrMSY1/OuCYR7j/Z
HQ68z9UD5ibxmvJ+tahzvIV48c0Bqn+m4hLkD4A+ly+KBXNBQLFleWM9LU6SnylR
SGbeYAPUWnrYKhL2bc+cFH/pnWXfD1Rr4QgDVGUALCSfc2uU9JvEe2nBaCGZv/jG
GGHrwl1DMxriHkc+Wp8R0luIVkZLOU0kp9xXF3H/gq6ca8qJ7rKhNtla+ntWneR8
Z4HMpfm5S4X2UvCrhdefxoqCntelT2YYK8e71qj+BG5M3pGsDrKaU8aF/9iz8vDV
nMOzL+Bga5U1HW6x6L5s9I/qT2iKnUofUP+UjJn+cPSjeltlsXpiYnEWEmKUKTcs
F6p3DVMj8CMiPnp0ywrVh632fD4DMXG/n35eX6xArzyUDPCzv0BF37y1LmFoDDRI
AF0tc1vY2fKt1+v+4jorakpckq2kXRK3PIsIxqAW5rzW9CjDTUzyRoh4eaMeKHxO
VCkWgGClC+su3Vrrj/JQRKDFoHoLazLyn9INk+leWqsvLzJYyoR3XSxq4wFQ9Exk
VZ2G3+ewMAj/74ujHCxHc0VviKYFDOwrK/FYcVBe7VEwVRgc1QNabi2IoBLYhbD4
m5bHlIJP1AScApDg7LHdEmR/jmwWYQIDAQABo1AwTjAdBgNVHQ4EFgQUv6x97x0+
6yIgaMbIr1DdipPazjgwHwYDVR0jBBgwFoAUv6x97x0+6yIgaMbIr1DdipPazjgw
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAgEAA6MmSixMgRpgEmfWODgy
qgyzBmwX0GlX/TVtc7SrOFAipoIcjTQdTEdLGLP8pX3FHo1sUok++gdZHkyfiPea
AOVZmCTk9CQ4mOi+xEKuQF7xfeyTpz2hXG06UZRRZpDX2F79WDLNzIGF0XRZxA3/
gp+q5iSGi2uNjTcFVAyIGysD9PXpfmMI40YZQtdHLUpLiRmdUPS7imJ+Rbau2JM8
ng1g3Yq2HTHcKFZggxAfWRBKiMf68MOjY/mU3sKd/Uv92dz0eZGOAhBJAoVQwhMb
WqZsegqstOpgVtt93LqqHm7ekDmzMXbSPonRST9zNJmy8WHof0/SNcz0/BxQjtB1
WznNyU2HGsOHzc650AQmsPhx6gjXRPsZiS0kqpAANW0M5i3vflE69XNnarhbJttQ
ruPhwlwLr96sL5kd24T5NmkDR3AV2qYURjM90Po+JZIDK31hzHm1puy0ZAymrNII
k0eyiBHzzlQsK5XSdhl74yELsgqs+lCxicNSAbYdXRer0F2HKuWIE5b0sanEgbUU
WvnssTAdeDWAeEVZr6+MYOlNk86npvIIfoKlpFehAcj8ZZ4LtTM/GFySn//oAwPx
aOgCM9yg3/8+LZ2QT0MKSciNwNQfG9Q5kmCq23UtAEdMYzgTtwHOSgXcFAQUOwmD
9X+8983a6+q5GxotSIQkF9w=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFKjCCAxICCQCGucd23XFo1jANBgkqhkiG9w0BAQsFADBVMQswCQYDVQQGEwJO
TDETMBEGA1UECAwKU29tZS1TdGF0ZTEVMBMGA1UECgwMT2ZmLVN5bmMuY29tMRow
GAYDVQQDDBF0ZXN0QG9mZi1zeW5jLmNvbTAeFw0xNzA3MjIxMTI4NDVaFw0yNzA3
MjAxMTI4NDVaMFkxCzAJBgNVBAYTAk5MMRMwEQYDVQQIDApTb21lLVN0YXRlMRUw
EwYDVQQKDAxPZmYtU3luYy5jb20xHjAcBgNVBAMMFXBsYXRmb3JtLXByb3h5LWRv
bWFpbjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAPMAGFsIg0MZ3+5p
jb10svnm08GX8ZyRlymYwTG2n9mczg9g9dZw6BpZM0ZzHNBgrewsQF265earBh3x
v1uLQwqFXHlkF9W22u2jAUK+KQ9dWYkbWnNQteCmPaNtAgojt5rBBdQEsZGTMe/p
ywFxzlgGgTUHLWpzMwa66zG+Aa3/bqkn4IIdfmocnH5lsXhsOFyxMNkh5KMgkMFQ
FzyqyV1lPPAF42ovbm/G7Zi8FJMdBljLFh9f33ONpHVfVuo4VqTihGOrV1PmAyd7
Aq6R/8I1aEy7vYXVGajUW57eMhnlGU/6ScjXlxnD1IP9E2c86nasbkRhHrNefd+E
S+SSKslm/I9j2vg9Sfwi/J8k3bKgTOwdN2w9VsqH2qft9I2CvsX5wfdgiKpu2DZ8
A3tcgKa1axgQ6upljOGLrkfXlOC2JVfEmVUjfReomjFOLbBUJ9zMdggnnYS9KsNH
YPSvmYuN1jc09iB0pWIdIBnWD13HbK4kSRf+YrtlUcE9QaULGFo0ypnI2iaVaNUv
ZcyNYOGkPWdtoq1x2q2/ef9dHsO1p1UDelfR9EnwKjb5VyyBaHJYNygA1UpOF0Ce
WaQ+4uYhb0E8rL3TjxX+2sdju3ZlaS/6Y8B2m2AkB2FhoPCuzOXlI6X32rMJ2yXs
rbOOrouuT7pf7QiYfsZAzgM75d9LAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAEfR
ylJZ8e16XPj9RmiEvsnfXhwoOBRRiGpV3077zlgwKfnZkIURtHTCOq73st1DRoTX
LPp9KjVibMJxGRm2ARwlpclkavvxDbN13gwjm2NxWOqs3IXKSnYxDEXn8PrUglqL
i2bbfT06udMreMx1tUOOhxDQDiJe5yGtlxLcocxeBBVLSK1evoyFixr5G7rkhaAy
YcdE4Zl5XnO8fXZ/4Nnr8SPXVKMglGw3YOiLIEgSCad5p2Z+VZ64apw2z0ipSNMy
ESTQtgimv75UY1GVLoBMoMwSgAXEq5wvLy18x1jNoawirWKmDMIWXxK3Rm3IbS6y
rTQCksdkucD6kMi1lgFUVDIUQAGbBZ8PG7uvY6NnahAk/TfN5OmxdYcalLmd4f9+
OOuRYHPhWapZ+8BueVggV3kQS5p+ktyixBFZRzkrw97CxwJaIjaQ6ZXji8HWyeqi
IgHwiDtwO4lQ+d++istUrPiG7DlaKS7o/2CYjvd6WFYcI+h79E+77HHaLKfdErEY
oPJ8BUcBBhuwDVJbYLiNlcvKYyhtmx86ciPWocoKwwjbljj4BeUw5HVDvAcO3YEv
+mbEpspWmpz/otoCiDLbHC96oKYxe3yeYmh8j5qzdS6F28unBBSYSvm4SXS+mPEF
4SxRj4PZCGJUpXLSUAMgL589zwUScalzg7RE27Ms
-----END CERTIFICATE-----
`

var keyPEM = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`
