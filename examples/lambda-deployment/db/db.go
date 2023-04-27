package db

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
	// Get the count.
	lsv, ok := gio.Item["count"]
	if !ok {
		return
	}
	// Check the type of the attribute.
	lsvs, ok := lsv.(*types.AttributeValueMemberN)
	if !ok {
		return
	}
	// Parse the attribute value.
	count, err = strconv.Atoi(lsvs.Value)
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
	// Get the count.
	lsv, ok := uio.Attributes["count"]
	if !ok {
		return
	}
	// Check the type of the attribute.
	lsvs, ok := lsv.(*types.AttributeValueMemberN)
	if !ok {
		return
	}
	// Parse the attribute value.
	count, err = strconv.Atoi(lsvs.Value)
	return
}
