package net

import (
	"fmt"
	"net/netip"

	"github.com/olxbr/network-api/pkg/types"
)

type NetworkNotInPoolError struct {
	Network *netip.Prefix
	Pool    *types.Pool
}

func (e NetworkNotInPoolError) Error() string {
	return fmt.Sprintf("network %s not in pool range %s", e.Network.String(), e.Pool.Range().String())
}
