package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynatypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/olxbr/network-api/pkg/types"
)

func (d *database) ScanPools(ctx context.Context) ([]*types.Pool, error) {
	paginator := dynamodb.NewScanPaginator(d.Client, &dynamodb.ScanInput{
		TableName: aws.String("napi_pools"),
	})

	pools := []*types.Pool{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return pools, err
		}

		var ps []*types.Pool
		err = attributevalue.UnmarshalListOfMaps(page.Items, &ps)
		if err != nil {
			return nil, err
		}
		pools = append(pools, ps...)
	}

	return pools, nil
}

func (d *database) GetPool(ctx context.Context, id string) (*types.Pool, error) {
	so, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("napi_pools"),
		Key: map[string]dynatypes.AttributeValue{
			"id": &dynatypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	pool := &types.Pool{}
	err = attributevalue.UnmarshalMap(so.Item, pool)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (d *database) PutPool(ctx context.Context, p *types.Pool) error {
	item, err := attributevalue.MarshalMap(p)
	if err != nil {
		return err
	}

	_, err = d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("napi_pools"),
		Item:      item,
	})
	return err
}

func (d *database) DeletePool(ctx context.Context, id string) error {
	_, err := d.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("napi_pools"),
		Key: map[string]dynatypes.AttributeValue{
			"id": &dynatypes.AttributeValueMemberS{Value: id},
		},
	})
	return err
}
