package db

import (
	"context"
	"errors"
	"fmt"

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
	qo, err := d.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("napi_pools"),
		KeyConditionExpression: aws.String("id = :hashKey"),
		ExpressionAttributeValues: map[string]dynatypes.AttributeValue{
			":hashKey": &dynatypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	if qo.Count <= 0 {
		return nil, errors.New("pool not found")
	}

	pool := &types.Pool{}
	err = attributevalue.UnmarshalMap(qo.Items[0], pool)
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

	item["sk"] = &dynatypes.AttributeValueMemberS{
		Value: fmt.Sprintf("%s#%s#%s", p.Region, p.SubnetIP, p.Name),
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
