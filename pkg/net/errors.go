package net

import (
	"fmt"

	"github.com/olxbr/network-api/pkg/types"
	"inet.af/netaddr"
)

type NetworkNotInPoolError struct {
	Network *netaddr.IPPrefix
	Pool    *types.Pool
}

func (e NetworkNotInPoolError) Error() string {
	return fmt.Sprintf("network %s not in pool range %s", e.Network.String(), e.Pool.Range().String())
}
