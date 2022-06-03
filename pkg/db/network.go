package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynatypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/olxbr/network-api/pkg/types"
)

func (d *database) ScanNetworks(ctx context.Context) ([]*types.Network, error) {
	paginator := dynamodb.NewScanPaginator(d.Client, &dynamodb.ScanInput{
		TableName: aws.String("napi_networks"),
	})

	networks := []*types.Network{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return networks, err
		}

		var ns []*types.Network
		err = attributevalue.UnmarshalListOfMaps(page.Items, &ns)
		if err != nil {
			return nil, err
		}
		networks = append(networks, ns...)
	}

	return networks, nil
}

func (d *database) GetNetwork(ctx context.Context, id string) (*types.Network, error) {
	so, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("napi_networks"),
		Key: map[string]dynatypes.AttributeValue{
			"id": &dynatypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	network := &types.Network{}
	err = attributevalue.UnmarshalMap(so.Item, network)
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (d *database) PutNetwork(ctx context.Context, n *types.Network) error {
	item, err := attributevalue.MarshalMap(n)
	if err != nil {
		return err
	}

	item["sk"] = &dynatypes.AttributeValueMemberS{
		Value: fmt.Sprintf("%s#%s#%s#%s#%s", n.Provider, n.Region, n.Account, n.Environment, n.CIDR),
	}

	_, err = d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("napi_networks"),
		Item:      item,
	})
	return err
}

func (d *database) DeleteNetwork(ctx context.Context, id string) error {
	_, err := d.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("napi_networks"),
		Key: map[string]dynatypes.AttributeValue{
			"id": &dynatypes.AttributeValueMemberS{Value: id},
		},
	})
	return err
}
