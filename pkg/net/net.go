package net

import (
	"context"
	"fmt"
	"log"
	"net/netip"

	"go4.org/netipx"

	"github.com/olxbr/network-api/pkg/db"
)

type NetworkManager struct {
	DB db.Database
}

func New(database db.Database) *NetworkManager {
	return &NetworkManager{DB: database}
}

func (nm *NetworkManager) CheckNetwork(ctx context.Context, network netip.Prefix) error {
	nets, err := nm.DB.ScanNetworks(ctx)
	if err != nil {
		return err
	}

	ipSetBuilder := &netipx.IPSetBuilder{}

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

func (nm *NetworkManager) AllocateNetwork(ctx context.Context, poolID string, subnetSize int) (netip.Prefix, error) {
	p, err := nm.DB.GetPool(ctx, poolID)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("error getting pool: %+v", err)
	}
	pn := p.Network()

	nets, err := nm.DB.ScanNetworks(ctx)
	if err != nil {
		return netip.Prefix{}, err
	}

	ipSetBuilder := &netipx.IPSetBuilder{}
	for _, n := range nets {
		ipSetBuilder.AddPrefix(n.IPPrefix())
	}

	ipset, err := ipSetBuilder.IPSet()
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("error building ipset: %+v", err)
	}

	newNet := netip.PrefixFrom(pn, subnetSize)

	for {
		if !ipset.OverlapsPrefix(newNet) {
			break
		}

		var valid bool
		newNet, valid = NextSubnet(newNet, subnetSize)
		if !valid {
			return netip.Prefix{}, fmt.Errorf("no more networks available")
		}

		newRange := netipx.RangeOfPrefix(newNet)
		if !p.Range().Overlaps(newRange) {
			return netip.Prefix{}, NetworkNotInPoolError{Network: &newNet, Pool: p}
		}
	}

	log.Printf("allocated network: %+v", newNet.String())
	return newNet, nil
}

func NextSubnet(n netip.Prefix, subnetSize int) (netip.Prefix, bool) {
	lastAddr := netipx.RangeOfPrefix(n).To()
	next := netip.PrefixFrom(lastAddr.Next(), subnetSize)
	return next, next.IsValid()
}
