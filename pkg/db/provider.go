package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynatypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/olxbr/network-api/pkg/types"
)

func (d *database) ScanProviders(ctx context.Context) ([]*types.Provider, error) {
	paginator := dynamodb.NewScanPaginator(d.Client, &dynamodb.ScanInput{
		TableName: aws.String("napi_providers"),
	})

	pools := []*types.Provider{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pools, err
		}

		var ps []*types.Provider
		err = attributevalue.UnmarshalListOfMaps(page.Items, &ps)
		if err != nil {
			return nil, err
		}
		pools = append(pools, ps...)
	}

	return pools, nil
}

func (d *database) GetProvider(ctx context.Context, region string) (*types.Provider, error) {
	so, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("napi_providers"),
		Key: map[string]dynatypes.AttributeValue{
			"name": &dynatypes.AttributeValueMemberS{Value: region},
		},
	})
	if err != nil {
		return nil, err
	}

	pool := &types.Provider{}
	err = attributevalue.UnmarshalMap(so.Item, pool)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (d *database) PutProvider(ctx context.Context, p *types.Provider) error {
	item, err := attributevalue.MarshalMap(p)
	if err != nil {
		return err
	}

	_, err = d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("napi_providers"),
		Item:      item,
	})
	return err
}

func (d *database) DeleteProvider(ctx context.Context, name string) error {
	_, err := d.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("napi_providers"),
		Key: map[string]dynatypes.AttributeValue{
			"name": &dynatypes.AttributeValueMemberS{Value: name},
		},
	})
	return err
}
