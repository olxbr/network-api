package net

import (
	"context"
	"fmt"
	"log"

	"inet.af/netaddr"

	"github.com/olxbr/network-api/pkg/db"
)

type NetworkManager struct {
	DB db.Database
}

func New(database db.Database) *NetworkManager {
	return &NetworkManager{DB: database}
}

func (nm *NetworkManager) CheckNetwork(ctx context.Context, network netaddr.IPPrefix) error {
	nets, err := nm.DB.ScanNetworks(ctx)
	if err != nil {
		return err
	}

	ipSetBuilder := &netaddr.IPSetBuilder{}

	for _, n := range nets {
		ipSetBuilder.AddPrefix(n.IPPrefix())
	}

	ipset, err := ipSetBuilder.IPSet()
	if err != nil {
		return fmt.Errorf("error building ipset: %+v", err)
	}

	if ipset.ContainsPrefix(network) {
		return fmt.Errorf("network %s overlaps with existing network", network.String())
	}
	return nil
}

func (nm *NetworkManager) AllocateNetwork(ctx context.Context, region string, subnetSize uint8) (netaddr.IPPrefix, error) {
	p, err := nm.DB.GetPool(ctx, region)
	if err != nil {
		return netaddr.IPPrefix{}, fmt.Errorf("error getting pool %s: %+v", region, err)
	}
	pn := p.Network()

	nets, err := nm.DB.ScanNetworks(ctx)
	if err != nil {
		return netaddr.IPPrefix{}, err
	}

	ipSetBuilder := &netaddr.IPSetBuilder{}
	for _, n := range nets {
		ipSetBuilder.AddPrefix(n.IPPrefix())
	}

	ipset, err := ipSetBuilder.IPSet()
	if err != nil {
		return netaddr.IPPrefix{}, fmt.Errorf("error building ipset: %+v", err)
	}

	newNet := netaddr.IPPrefixFrom(pn, subnetSize)

	for {
		if !ipset.OverlapsPrefix(newNet) {
			break
		}

		var valid bool
		newNet, valid = NextSubnet(newNet, subnetSize)
		if !valid {
			return netaddr.IPPrefix{}, fmt.Errorf("no more networks available")
		}

		if !p.Range().Overlaps(newNet.Range()) {
			return netaddr.IPPrefix{}, NetworkNotInPoolError{Network: &newNet, Pool: p}
		}
	}

	log.Printf("allocated network: %+v", newNet.String())
	return newNet, nil
}

func NextSubnet(n netaddr.IPPrefix, subnetSize uint8) (netaddr.IPPrefix, bool) {
	lastAddr := n.Range().To()
	next := netaddr.IPPrefixFrom(lastAddr.Next(), subnetSize)
	return next, next.IsValid()
}
