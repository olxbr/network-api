package db

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/olxbr/network-api/pkg/db/fake"
	"github.com/olxbr/network-api/pkg/types"
)

func TestCanCreateNetwork(t *testing.T) {
	cli := &fake.DynamoClient{}

	cli.On("PutItem", mock.Anything, mock.MatchedBy(func(params *dynamodb.PutItemInput) bool {
		id := fmt.Sprintf("%v", params.Item["id"])
		return aws.ToString(params.TableName) == "napi_networks" &&
			strings.Contains(id, "f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0")
	})).Return(&dynamodb.PutItemOutput{}, nil)

	d := New(cli)
	err := d.PutNetwork(context.TODO(), &types.Network{
		ID:   &types.DynamoUUID{UUID: uuid.MustParse("f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0")},
		CIDR: "10.0.0.0/24",
	})

	assert.NoError(t, err)
}
