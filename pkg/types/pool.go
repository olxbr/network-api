package types

import (
	"inet.af/netaddr"
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

type PoolResponse struct {
	Items []*Pool `json:"items"`
}

func (p Pool) Network() netaddr.IP {
	return netaddr.MustParseIP(p.SubnetIP)
}

func (p Pool) Range() netaddr.IPRange {
	ip := netaddr.MustParseIP(p.SubnetIP)
	if p.SubnetMask != nil {
		ipnet := netaddr.IPPrefixFrom(ip, uint8(*p.SubnetMask))
		return ipnet.Range()
	}
	maxIP := netaddr.MustParseIP(*p.SubnetMaxIP)
	return netaddr.IPRangeFrom(ip, maxIP)
}
