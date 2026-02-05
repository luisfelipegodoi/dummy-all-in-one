package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoClient wraps AWS SDK DynamoDB client configured for LocalStack.
type DynamoClient struct {
	Client *dynamodb.Client
}

func NewDynamoDB(ctx context.Context, region, endpoint string) (*DynamoClient, error) {
	if region == "" {
		region = "sa-east-1"
	}
	if endpoint == "" {
		return nil, errors.New("dynamodb endpoint is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
	if err != nil {
		return nil, err
	}

	return &DynamoClient{Client: dynamodb.NewFromConfig(cfg)}, nil
}

// WaitTableStatus polls DescribeTable until it matches the desired status.
func (c *DynamoClient) WaitTableStatus(ctx context.Context, table string, want types.TableStatus, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	deadline := time.Now().Add(timeout)
	backoff := 300 * time.Millisecond

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting table %q status=%s", table, want)
		}

		out, err := c.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		})
		if err == nil && out.Table != nil && out.Table.TableStatus == want {
			return nil
		}

		time.Sleep(backoff)
		if backoff < 2*time.Second {
			backoff *= 2
		}
	}
}

// ListTables returns all tables (paginado).
func (c *DynamoClient) ListTables(ctx context.Context) ([]string, error) {
	var (
		all      []string
		startKey *string
	)

	for {
		out, err := c.Client.ListTables(ctx, &dynamodb.ListTablesInput{
			ExclusiveStartTableName: startKey,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, out.TableNames...)

		if out.LastEvaluatedTableName == nil || *out.LastEvaluatedTableName == "" {
			break
		}
		startKey = out.LastEvaluatedTableName
	}

	return all, nil
}

// AssertTablesExist waits until all expected tables exist (via DescribeTable).
func (c *DynamoClient) AssertTablesExist(ctx context.Context, tables []string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	deadline := time.Now().Add(timeout)

	missing := func() []string {
		var m []string
		for _, t := range tables {
			_, err := c.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(t)})
			if err != nil {
				var notFound *types.ResourceNotFoundException
				if errors.As(err, &notFound) {
					m = append(m, t)
					continue
				}
				// other error: fail fast
				return []string{fmt.Sprintf("%s (error: %v)", t, err)}
			}
		}
		return m
	}

	for {
		m := missing()
		if len(m) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting tables exist. missing=%v", m)
		}
		time.Sleep(1 * time.Second)
	}
}

// ----- "Editar tabela" (exemplos úteis) -----

// UpdateTTL enables/disables TTL for a table.
func (c *DynamoClient) UpdateTTL(ctx context.Context, table, attrName string, enabled bool) error {
	_, err := c.Client.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(table),
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			AttributeName: aws.String(attrName),
			Enabled:       aws.Bool(enabled),
		},
	})
	return err
}

// PutItemSimple convenience helper.
func (c *DynamoClient) PutItem(ctx context.Context, table string, item map[string]types.AttributeValue) error {
	_, err := c.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	return err
}

// GetItemSimple convenience helper.
func (c *DynamoClient) GetItem(ctx context.Context, table string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	out, err := c.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(table),
		Key:            key,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return out.Item, nil
}

// WaitForStatusChange compares an item attribute (e.g. "status") until it becomes expected.
// Use this to "compare status" and confirm it changed.
//
// Example:
//
//	key := map[string]types.AttributeValue{"pk": &types.AttributeValueMemberS{Value: "123"}}
//	err := dyn.WaitForStatusChange(ctx, "table1", key, "status", "DONE", 30*time.Second)
func (c *DynamoClient) WaitForStatusChange(
	ctx context.Context,
	table string,
	key map[string]types.AttributeValue,
	statusAttr string,
	expected string,
	timeout time.Duration,
) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	deadline := time.Now().Add(timeout)
	backoff := 300 * time.Millisecond

	for {
		if time.Now().After(deadline) {
			cur, _ := c.readStatusOnce(ctx, table, key, statusAttr)
			return fmt.Errorf("timeout waiting status=%q in %s. current=%q", expected, statusAttr, cur)
		}

		cur, err := c.readStatusOnce(ctx, table, key, statusAttr)
		if err == nil && cur == expected {
			return nil
		}

		time.Sleep(backoff)
		if backoff < 2*time.Second {
			backoff *= 2
		}
	}
}

func (c *DynamoClient) readStatusOnce(
	ctx context.Context,
	table string,
	key map[string]types.AttributeValue,
	statusAttr string,
) (string, error) {
	out, err := c.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(table),
		Key:            key,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	if len(out.Item) == 0 {
		return "", fmt.Errorf("item not found")
	}

	av, ok := out.Item[statusAttr]
	if !ok {
		return "", fmt.Errorf("status attribute %q not found in item", statusAttr)
	}

	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return v.Value, nil
	default:
		return "", fmt.Errorf("status attribute %q is not string (got %T)", statusAttr, av)
	}
}
