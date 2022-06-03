package types

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynatypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamoUUID struct {
	uuid.UUID
}

func NewUUID() *DynamoUUID {
	return &DynamoUUID{uuid.New()}
}

func (u *DynamoUUID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avS, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return nil
	}

	n, err := uuid.Parse(avS.Value)
	if err != nil {
		return err
	}

	u.UUID = n
	return nil
}

func (u *DynamoUUID) MarshalDynamoDBAttributeValue() (dynatypes.AttributeValue, error) {
	return &types.AttributeValueMemberS{
		Value: u.UUID.String(),
	}, nil
}
