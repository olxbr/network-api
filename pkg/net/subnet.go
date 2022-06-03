package net

import (
	"fmt"
	"log"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/olxbr/network-api/pkg/types"
)

func GenerateSubnets(n *types.Network) ([]*types.Subnet, error) {
	snets := []*types.Subnet{}

	_, baseCIDR, err := net.ParseCIDR(n.CIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid base CIDR: %w", err)
	}

	azSubnets := make([]*net.IPNet, 4)

	ones, _ := baseCIDR.Mask.Size()
	subnetSize := ones + 2
	log.Printf("subnetSize: %d", subnetSize)

	for i := 0; i < len(azSubnets); i++ {
		newSubnet, err := cidr.Subnet(baseCIDR, 2, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create subnet: %w", err)
		}
		fmt.Printf("%s\n", newSubnet.String())
		azSubnets[i] = newSubnet
	}

	if n.PrivateSubnet {
		subnetSize += 1
		for i := 0; i < len(azSubnets)-1; i++ {
			snet, err := cidr.Subnet(azSubnets[i], 1, 0)
			if err != nil {
				log.Printf("failed to create subnet: %s", err)
			}
			snets = append(snets, &types.Subnet{
				Name: fmt.Sprintf("private%02d", i+1),
				Type: types.Private,
				CIDR: snet.String(),
			})
			azSubnets[i], _ = cidr.NextSubnet(snet, subnetSize)
		}
	}

	if n.PublicSubnet {
		subnetSize += 1
		for i := 0; i < len(azSubnets)-1; i++ {
			snet, err := cidr.Subnet(azSubnets[i], 1, 0)
			if err != nil {
				log.Printf("failed to create subnet: %s", err)
			}
			snets = append(snets, &types.Subnet{
				Name: fmt.Sprintf("public%02d", i+1),
				Type: types.Public,
				CIDR: snet.String(),
			})
			azSubnets[i], _ = cidr.NextSubnet(snet, subnetSize)
		}
	}

	if n.AttachTGW {
		log.Printf("subnet size: %d\n", subnetSize)
		if subnetSize > 28 {
			return nil, fmt.Errorf("cannot attach TGW to subnet with size < 28: increase base CIDR range")
		}
		for i := 0; i < len(azSubnets)-1; i++ {
			snet, _ := cidr.Subnet(azSubnets[i], 28-subnetSize, 0)
			snets = append(snets, &types.Subnet{
				Name: fmt.Sprintf("tgw%02d", i+1),
				Type: types.TransitGateway,
				CIDR: snet.String(),
			})
		}
	}
	return snets, nil
}
