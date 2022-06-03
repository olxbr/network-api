package types

type Provider struct {
	ID         *DynamoUUID `json:"id" dynamodbav:"id"`
	Name       string      `json:"name" dynamodbav:"name"`
	WebhookURL string      `json:"webhookURL" dynamodbav:"webhookURL"`
	APIToken   string      `json:"apiToken" dynamodbav:"apiToken"`
}

type ProviderRequest struct {
	Name       string `json:"name" validate:"required"`
	WebhookURL string `json:"webhookURL" validate:"required"`
	APIToken   string `json:"apiToken" validate:"required"`
}

type ProviderResponse struct {
	Items []*Provider `json:"items"`
}

type ProviderUpdateRequest struct {
	WebhookURL *string `json:"vpcID,omitempty"`
	APIToken   *string `json:"info,omitempty"`
}

type EventType string

const (
	CreateNetwork EventType = "create_network"
	CheckNework   EventType = "check_network"
	DeleteNetwork EventType = "delete_network"
	QueryNetwork  EventType = "query_network"
)

type ProviderWebhook struct {
	Event       EventType `json:"event"`
	NetworkID   string    `json:"networkID" validate:"required"`
	Account     string    `json:"account" validate:"required"`
	Region      string    `json:"region" validate:"required"`
	Environment string    `json:"environment" validate:"required"`
	CIDR        string    `json:"cidr" validate:"required_if=Event create_network,omitempty,cidr"`
	Subnets     []*Subnet `json:"subnets,omitempty" validate:"required_if=Event create_network,omitempty"`
}

type ProviderWebhookResponse struct {
	StatusCode int    `json:"statusCode"`
	ID         string `json:"id"`
}
