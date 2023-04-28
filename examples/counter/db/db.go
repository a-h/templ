package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type OptionsFunc func(*CountStore)

func WithClient(client *dynamodb.Client) func(*CountStore) {
	return func(ms *CountStore) {
		ms.db = client
	}
}

func NewCountStore(tableName, region string, options ...OptionsFunc) (s *CountStore, err error) {
	s = &CountStore{
		tableName: tableName,
	}
	for _, o := range options {
		o(s)
	}
	if s.db == nil {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
		if err != nil {
			return s, err
		}
		s.db = dynamodb.NewFromConfig(cfg)
	}
	return
}

type CountStore struct {
	db        *dynamodb.Client
	tableName string
}

func stripEmpty(strings []string) (op []string) {
	for _, s := range strings {
		if s != "" {
			op = append(op, s)
		}
	}
	return
}

type countRecord struct {
	PK    string `dynamodbav:"_pk"`
	Count int    `dynamodbav:"count"`
}

func (s CountStore) BatchGet(ctx context.Context, ids ...string) (counts []int, err error) {
	nonEmptyIDs := stripEmpty(ids)
	if len(nonEmptyIDs) == 0 {
		return nil, nil
	}

	// Make DynamoDB keys.
	ris := make(map[string]types.KeysAndAttributes)
	for _, id := range nonEmptyIDs {
		ri := ris[s.tableName]
		ri.Keys = append(ris[s.tableName].Keys, map[string]types.AttributeValue{
			"_pk": &types.AttributeValueMemberS{
				Value: id,
			},
		})
		ri.ConsistentRead = aws.Bool(true)
		ris[s.tableName] = ri
	}

	// Execute the batch request.
	var batchResponses []map[string]types.AttributeValue

	// DynamoDB might not process everything, so we need a loop.
	var unprocessedAttempts int
	for {
		var bgio *dynamodb.BatchGetItemOutput
		bgio, err = s.db.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: ris,
		})
		if err != nil {
			return
		}
		for _, responses := range bgio.Responses {
			batchResponses = append(batchResponses, responses...)
		}
		if len(bgio.UnprocessedKeys) > 0 {
			ris = bgio.UnprocessedKeys
			unprocessedAttempts++
			if unprocessedAttempts > 3 {
				err = fmt.Errorf("countstore: exceeded three attempts to get all counts")
				return
			}
			continue
		}
		break
	}

	// Process the responses into structs.
	crs := []countRecord{}
	err = attributevalue.UnmarshalListOfMaps(batchResponses, &crs)
	if err != nil {
		err = fmt.Errorf("countstore: failed to unmarshal result of BatchGet: %w", err)
		return
	}

	// Match up the inputs to the records.
	idToCount := make(map[string]int, len(ids))
	for _, cr := range crs {
		idToCount[cr.PK] = cr.Count
	}

	// Create the output in the right order.
	// Missing values are defaulted to zero.
	for _, id := range ids {
		counts = append(counts, idToCount[id])
	}

	return
}

func (s CountStore) Get(ctx context.Context, id string) (count int, err error) {
	if id == "" {
		return
	}
	gio, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"_pk": &types.AttributeValueMemberS{
				Value: id,
			},
		},
		TableName:      &s.tableName,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil || gio.Item == nil {
		return
	}

	var cr countRecord
	err = attributevalue.UnmarshalMap(gio.Item, &cr)
	if err != nil {
		return 0, fmt.Errorf("countstore: failed to process result of Get: %w", err)
	}
	count = cr.Count

	return
}

func (s CountStore) Increment(ctx context.Context, id string) (count int, err error) {
	if id == "" {
		return
	}
	uio, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"_pk": &types.AttributeValueMemberS{
				Value: id,
			},
		},
		TableName:        &s.tableName,
		UpdateExpression: aws.String("SET #c = if_not_exists(#c, :zero) + :one"),
		ExpressionAttributeNames: map[string]string{
			"#c": "count",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
			":one":  &types.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return
	}

	// Parse the response.
	var cr countRecord
	err = attributevalue.UnmarshalMap(uio.Attributes, &cr)
	if err != nil {
		return 0, fmt.Errorf("countstore: failed to process result of Increment: %w", err)
	}
	count = cr.Count

	return
}
