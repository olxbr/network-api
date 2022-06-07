package aws

import (
	"testing"

	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildParameters(t *testing.T) {

	params := BuildParameters(&types.ProviderWebhook{
		NetworkID:   "f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0",
		CIDR:        "10.1.0.0/16",
		Environment: "prod",
		Subnets: []*types.Subnet{
			{Name: "private-0", Type: types.Private, CIDR: "10.1.0.0/19"},
			{Name: "private-1", Type: types.Private, CIDR: "10.1.64.0/19"},
			{Name: "private-2", Type: types.Private, CIDR: "10.1.128.0/19"},
			{Name: "public-0", Type: types.Public, CIDR: "10.1.32.0/20"},
			{Name: "public-1", Type: types.Public, CIDR: "10.1.96.0/20"},
			{Name: "public-2", Type: types.Public, CIDR: "10.1.160.0/20"},
			{Name: "attach-0", Type: types.TransitGateway, CIDR: "10.1.48.0/28"},
			{Name: "attach-1", Type: types.TransitGateway, CIDR: "10.1.112.0/28"},
			{Name: "attach-2", Type: types.TransitGateway, CIDR: "10.1.176.0/28"},
		},
	})

	assert.Equal(t, "f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0", *params[0].ParameterValue)
}
