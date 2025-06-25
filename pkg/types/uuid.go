package types

import (
	dynatypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamoUUID struct {
	uuid.UUID
}

func NewUUID() *DynamoUUID {
	return &DynamoUUID{uuid.New()}
}

func (u *DynamoUUID) UnmarshalDynamoDBAttributeValue(av dynatypes.AttributeValue) error {
	avS, ok := av.(*dynatypes.AttributeValueMemberS)
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
	return &dynatypes.AttributeValueMemberS{
		Value: u.String(),
	}, nil
}
