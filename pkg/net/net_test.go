package net

import (
	"context"
	"log"
	"testing"

	"github.com/olxbr/network-api/pkg/db/fake"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"inet.af/netaddr"
)

func TestAllocateNetwork(t *testing.T) {
	tests := []struct {
		name       string
		poolID     string
		subnetSize uint8
		prepare    func(t *testing.T, db *fake.Database)
		assert     func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error)
	}{
		{
			name:       "empty database",
			poolID:     "poolid",
			subnetSize: 24,
			prepare: func(t *testing.T, db *fake.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
			},
			assert: func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error) {
				db.AssertExpectations(t)
				assert.NoError(t, err)
				assert.Equal(t, "10.0.0.0/24", n.String())
			},
		},
		{
			name:       "with existing networks",
			poolID:     "poolid",
			subnetSize: 24,
			prepare: func(t *testing.T, db *fake.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{
					{CIDR: "10.0.0.0/24"},
					{CIDR: "10.0.1.0/24"},
				}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
			},
			assert: func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error) {
				db.AssertExpectations(t)
				assert.NoError(t, err)
				assert.Equal(t, "10.0.2.0/24", n.String())
			},
		},
		{
			name:       "with existing networks 2",
			poolID:     "poolid",
			subnetSize: 24,
			prepare: func(t *testing.T, db *fake.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{
					{CIDR: "10.0.0.0/16"},
					{CIDR: "10.1.0.0/24"},
				}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
			},
			assert: func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error) {
				db.AssertExpectations(t)
				assert.NoError(t, err)
				assert.Equal(t, "10.1.1.0/24", n.String())
			},
		},
		{
			name:       "with existing different sizes networks",
			poolID:     "poolid",
			subnetSize: 23,
			prepare: func(t *testing.T, db *fake.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{
					{CIDR: "10.0.0.0/24"},
					{CIDR: "10.0.2.0/24"},
				}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
			},
			assert: func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error) {
				db.AssertExpectations(t)
				assert.NoError(t, err)
				assert.Equal(t, "10.0.4.0/23", n.String())
			},
		},
		{
			name:       "result network is not in pool",
			poolID:     "poolid",
			subnetSize: 10,
			prepare: func(t *testing.T, db *fake.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{
					{CIDR: "10.0.0.0/24"},
					{CIDR: "10.64.0.0/10"},
					{CIDR: "10.128.0.0/24"},
					{CIDR: "10.192.0.0/23"},
				}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
			},
			assert: func(t *testing.T, db *fake.Database, n netaddr.IPPrefix, err error) {
				db.AssertExpectations(t)
				assert.ErrorContains(t, err, "network 11.0.0.0/10 not in pool range 10.0.0.0-10.255.255.255")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("Running test: %s", tt.name)
			db := &fake.Database{}
			nm := New(db)
			ctx := context.Background()

			tt.prepare(t, db)
			net, err := nm.AllocateNetwork(ctx, tt.poolID, tt.subnetSize)
			tt.assert(t, db, net, err)
		})
	}

}

func TestNextSubnet(t *testing.T) {
	tests := []struct {
		name     string
		previous netaddr.IPPrefix
		next     netaddr.IPPrefix
		valid    bool
	}{
		{
			name:     "9.255.255.0/24",
			previous: netaddr.MustParseIPPrefix("9.255.255.0/24"),
			next:     netaddr.MustParseIPPrefix("10.0.0.0/24"),
			valid:    true,
		},
		{
			name:     "99.255.255.192/26",
			previous: netaddr.MustParseIPPrefix("99.255.255.192/26"),
			next:     netaddr.MustParseIPPrefix("100.0.0.0/26"),
			valid:    true,
		},
		{
			name:     "255.255.255.192/26",
			previous: netaddr.MustParseIPPrefix("255.255.255.192/26"),
			next:     netaddr.MustParseIPPrefix("0.0.0.0/26"),
			valid:    false,
		},
		{
			name:     "2001:db8:d000::/36",
			previous: netaddr.MustParseIPPrefix("2001:db8:d000::/36"),
			next:     netaddr.MustParseIPPrefix("2001:db8:e000::/36"),
			valid:    true,
		},
		{
			name:     "ffff:ffff:ffff:ffff::/64",
			previous: netaddr.MustParseIPPrefix("ffff:ffff:ffff:ffff::/64"),
			next:     netaddr.MustParseIPPrefix("::/64"),
			valid:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("Running test: %s", tt.name)
			next, v := NextSubnet(tt.previous, tt.previous.Bits())
			log.Printf("next: %s", next)

			assert.Equal(t, tt.valid, v)
			if tt.valid {
				assert.Equal(t, tt.next, next)
			}
		})
	}
}
