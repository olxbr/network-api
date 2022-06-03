package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/olxbr/network-api/pkg/types"
)

type Database interface {
	ScanNetworks(ctx context.Context) ([]*types.Network, error)
	GetNetwork(ctx context.Context, id string) (*types.Network, error)
	PutNetwork(ctx context.Context, n *types.Network) error
	DeleteNetwork(ctx context.Context, id string) error

	ScanPools(ctx context.Context) ([]*types.Pool, error)
	GetPool(ctx context.Context, region string) (*types.Pool, error)
	PutPool(ctx context.Context, p *types.Pool) error
	DeletePool(ctx context.Context, region string) error

	ScanProviders(ctx context.Context) ([]*types.Provider, error)
	GetProvider(ctx context.Context, name string) (*types.Provider, error)
	PutProvider(ctx context.Context, p *types.Provider) error
	DeleteProvider(ctx context.Context, name string) error
}

type DynamoClient interface {
	dynamodb.ScanAPIClient
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type database struct {
	Client DynamoClient
}

func New(cli DynamoClient) Database {
	return &database{
		Client: cli,
	}
}
