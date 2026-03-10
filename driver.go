package godynamo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/btnguyen2k/consu/reddo"
)

// init is automatically invoked when the driver is imported
func init() {
	sql.Register("godynamo", &Driver{})
}

// Driver is AWS DynamoDB implementation of driver.Driver.
type Driver struct {
}

func parseParamValue(params map[string]string, typ reflect.Type, validator func(val interface{}) bool,
	defaultVal interface{}, pkeys []string, ekeys []string) interface{} {
	for _, key := range pkeys {
		val, ok := params[key]
		if ok {
			pval, err := reddo.Convert(val, typ)
			if pval == nil || err != nil || (validator != nil && !validator(pval)) {
				return defaultVal
			}
			return pval
		}
	}
	for _, key := range ekeys {
		val := os.Getenv(key)
		if val != "" {
			pval, err := reddo.Convert(val, typ)
			if pval == nil || err != nil || (validator != nil && !validator(pval)) {
				return defaultVal
			}
			return pval
		}
	}
	return defaultVal
}

func parseConnString(connStr string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		tokens := strings.SplitN(strings.TrimSpace(part), "=", 2)
		key := strings.ToUpper(strings.TrimSpace(tokens[0]))
		if len(tokens) >= 2 {
			params[key] = strings.TrimSpace(tokens[1])
		} else {
			params[key] = ""
		}
	}
	return params
}

// openConn creates a driver.Conn from a DSN string and an optional aws.Config.
// If cfg is non-nil, it is used to create the DynamoDB client (with DSN options merged in).
// Otherwise, static credentials from the DSN are used directly.
func openConn(connStr string, cfg *aws.Config) (driver.Conn, error) {
	params := parseConnString(connStr)
	timeoutMs := parseParamValue(params, reddo.TypeInt, func(val interface{}) bool {
		return val.(int64) >= 0
	}, int64(10000), []string{"TIMEOUTMS"}, nil).(int64)
	region := parseParamValue(params, reddo.TypeString, nil, "", []string{"REGION"}, []string{"AWS_REGION"}).(string)
	akid := parseParamValue(params, reddo.TypeString, nil, "", []string{"AKID"}, []string{"AWS_ACCESS_KEY_ID", "AWS_AKID"}).(string)
	secretKey := parseParamValue(params, reddo.TypeString, nil, "", []string{"SECRET_KEY", "SECRETKEY"}, []string{"AWS_SECRET_KEY", "AWS_SECRET_ACCESS_KEY"}).(string)
	opts := dynamodb.Options{
		Credentials: credentials.NewStaticCredentialsProvider(akid, secretKey, ""),
		HTTPClient:  http.NewBuildableClient().WithTimeout(time.Millisecond * time.Duration(timeoutMs)),
		Region:      region,
	}
	endpoint := parseParamValue(params, reddo.TypeString, nil, "", []string{"ENDPOINT"}, []string{"AWS_DYNAMODB_ENDPOINT"}).(string)
	if endpoint != "" {
		opts.BaseEndpoint = aws.String(endpoint)
		if strings.HasPrefix(endpoint, "http://") {
			opts.EndpointOptions.DisableHTTPS = true
		}
	}
	client := dynamodb.New(opts)

	if cfg != nil {
		client = dynamodb.NewFromConfig(*cfg, mergeDynamoDBOptions(opts))
	}

	return &Conn{client: client, timeout: time.Duration(timeoutMs) * time.Millisecond}, nil
}

// Open implements driver.Driver/Open.
//
// connStr is expected in the following format:
//
//	Region=<region>;AkId=<aws-key-id>;Secret_Key=<aws-secret-key>[;Endpoint=<dynamodb-endpoint>][;TimeoutMs=<timeout-in-milliseconds>]
//
// If not supplied, default value for TimeoutMs is 10 seconds.
//
// Open uses only the credentials provided in the connection string.
// To use an aws.Config (e.g. for shared credentials), use NewConnector with sql.OpenDB instead.
func (d *Driver) Open(connStr string) (driver.Conn, error) {
	return openConn(connStr, nil)
}

// Connector implements database/sql/driver.Connector for per-instance AWS configuration.
// Use NewConnector and sql.OpenDB to create connections without global state.
type Connector struct {
	dsn       string
	awsConfig *aws.Config
}

// NewConnector creates a Connector that holds per-instance AWS configuration.
// Use sql.OpenDB(connector) instead of sql.Open("godynamo", dsn) to avoid global state.
// If cfg is nil, static credentials from the DSN are used.
func NewConnector(dsn string, cfg *aws.Config) *Connector {
	return &Connector{dsn: dsn, awsConfig: cfg}
}

// Connect implements driver.Connector/Connect.
func (c *Connector) Connect(_ context.Context) (driver.Conn, error) {
	return openConn(c.dsn, c.awsConfig)
}

// Driver implements driver.Connector/Driver.
func (c *Connector) Driver() driver.Driver {
	return &Driver{}
}

// mergeDynamoDBOptions merges the provided dynamodb.Options into the default dynamodb.Options.
func mergeDynamoDBOptions(providedOpts dynamodb.Options) func(*dynamodb.Options) {
	return func(defaultOpts *dynamodb.Options) {
		if defaultOpts.Region == "" {
			defaultOpts.Region = providedOpts.Region
		}
		if defaultOpts.Credentials == nil {
			defaultOpts.Credentials = providedOpts.Credentials
		}
		defaultOpts.HTTPClient = providedOpts.HTTPClient

		if defaultOpts.BaseEndpoint == nil {
			defaultOpts.BaseEndpoint = providedOpts.BaseEndpoint
			defaultOpts.EndpointOptions = providedOpts.EndpointOptions
		}
	}
}
