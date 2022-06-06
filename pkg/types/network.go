package types

import (
	"fmt"
	"net"

	"inet.af/netaddr"
)

// SortKey: [Provider]#[Region]#[Account]#[Environment]#[CIDR]
type Network struct {
	ID          *DynamoUUID `json:"id" dynamodbav:"id"`
	Provider    string      `json:"provider" dynamodbav:"provider"`
	Region      string      `json:"region" dynamodbav:"region"`
	Account     string      `json:"account" dynamodbav:"account"`
	Environment string      `json:"environment" dynamodbav:"environment"`
	CIDR        string      `json:"cidr" dynamodbav:"cidr"`

	VpcID string `json:"vpcID" dynamodbav:"vpcID"`
	Info  string `json:"info" dynamodbav:"info"`

	AttachTGW     bool `json:"attachTGW,omitempty" dynamodbav:"attachTGW"`
	PrivateSubnet bool `json:"privateSubnet,omitempty" dynamodbav:"privateSubnet"`
	PublicSubnet  bool `json:"publicSubnet,omitempty" dynamodbav:"publicSubnet"`
	Legacy        bool `json:"legacy,omitempty" dynamodbav:"legacy"`
	Reserved      bool `json:"reserved,omitempty" dynamodbav:"reserved"`
}

type NetworkRequest struct {
	Account     string `json:"account" validate:"required"`
	Region      string `json:"region" validate:"required"`
	Provider    string `json:"provider" validate:"required"`
	Environment string `json:"environment" validate:"required"`

	Info string `json:"info,omitempty" validate:"omitempty"`

	SubnetSize int `json:"subnetSize" validate:"required_without_all=Reserved Legacy,omitempty,max=24,min=16"`

	AttachTGW     *bool `json:"attachTGW,omitempty" validate:"required"`
	PrivateSubnet *bool `json:"privateSubnet,omitempty" validate:"required"`
	PublicSubnet  *bool `json:"publicSubnet,omitempty" validate:"required"`
	Legacy        *bool `json:"legacy,omitempty" validate:"omitempty"`

	Reserved *bool  `json:"reserved,omitempty" validate:"omitempty"`
	CIDR     string `json:"cidr,omitempty" validate:"excluded_without_all=Reserved Legacy,omitempty,cidr"`
}

type NetworkResponse struct {
	Network *Network                 `json:"network"`
	Webhook *ProviderWebhookResponse `json:"webhook,omitempty"`
}
type NetworkUpdateRequest struct {
	VpcID *string `json:"vpcID,omitempty"`
	Info  *string `json:"info,omitempty"`
}

type NetworkListResponse struct {
	Items []*Network `json:"items"`
}

type SubnetResponse struct {
	Subnets []*Subnet `json:"subnets"`
}

type SubnetType string

const (
	Private        SubnetType = "private"
	Public         SubnetType = "public"
	TransitGateway SubnetType = "transitGateway"
)

type Subnet struct {
	Name string     `json:"name"`
	Type SubnetType `json:"type"`
	CIDR string     `json:"cidr"`
}

func (n Network) Network() net.IPNet {
	_, cidr, _ := net.ParseCIDR(n.CIDR)
	return *cidr
}

func (n Network) IPPrefix() netaddr.IPPrefix {
	return netaddr.MustParseIPPrefix(n.CIDR)
}

func (n Network) String() string {
	return fmt.Sprintf("<CIDR: %s, Account: %s, Region: %s>", n.CIDR, n.Account, n.Region)
}
