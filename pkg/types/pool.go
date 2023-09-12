package types

import (
	"net/netip"

	"go4.org/netipx"
)

type Pool struct {
	ID          *DynamoUUID `json:"id" dynamodbav:"id"`
	Name        string      `json:"name" dynamodbav:"name"`
	Region      string      `json:"region" dynamodbav:"region"`
	SubnetIP    string      `json:"subnetIP" dynamodbav:"cidr"`
	SubnetMask  *int        `json:"subnetMask,omitempty" dynamodbav:"subnetMask"`
	SubnetMaxIP *string     `json:"subnetMaxIP,omitempty" dynamodbav:"subnetMaxIP"`
}

type PoolRequest struct {
	Name        string  `json:"name" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	SubnetIP    string  `json:"subnetIP" validate:"required,ip"`
	SubnetMask  *int    `json:"subnetMask,omitempty" validate:"omitempty,max=24,min=8"`
	SubnetMaxIP *string `json:"subnetMaxIP,omitempty" validate:"required_without=SubnetMask,excluded_with=SubnetMask,omitempty,ip"`
}

type PoolListResponse struct {
	Items []*Pool `json:"items"`
}

func (p Pool) Network() netip.Addr {
	return netip.MustParseAddr(p.SubnetIP)
}

func (p Pool) Range() netipx.IPRange {
	ip := netip.MustParseAddr(p.SubnetIP)
	if p.SubnetMask != nil {
		prefix := netip.PrefixFrom(ip, int(*p.SubnetMask))
		return netipx.RangeOfPrefix(prefix)
	}
	maxIP := netip.MustParseAddr(*p.SubnetMaxIP)
	return netipx.IPRangeFrom(ip, maxIP)
}
