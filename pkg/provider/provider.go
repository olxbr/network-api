package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/olxbr/network-api/pkg/db"
	"github.com/olxbr/network-api/pkg/net"
	"github.com/olxbr/network-api/pkg/secret"
	"github.com/olxbr/network-api/pkg/types"
)

type ProviderManager struct {
	d db.Database
	s secret.Secrets
}

type ProviderClient struct {
	cli  http.Client
	auth string
	url  string
}

func New(database db.Database, s secret.Secrets) *ProviderManager {
	return &ProviderManager{
		d: database,
		s: s,
	}
}

func (p *ProviderManager) GetClient(ctx context.Context, name string) (*ProviderClient, error) {
	provider, err := p.d.GetProvider(ctx, name)
	if err != nil {
		return nil, err
	}

	token, err := p.s.GetAPIToken(ctx, name)
	if err != nil {
		return nil, err
	}

	return &ProviderClient{
		cli: http.Client{
			Timeout: 300 * time.Second,
		},
		auth: token,
		url:  provider.WebhookURL,
	}, nil
}

func (p *ProviderClient) CreateNetwork(ctx context.Context, n *types.Network) (*types.ProviderWebhookResponse, error) {
	subnets, err := net.GenerateSubnets(n)
	if err != nil {
		return nil, err
	}

	webhook := types.ProviderWebhook{
		Event:       types.CreateNetwork,
		NetworkID:   n.ID.String(),
		CIDR:        n.CIDR,
		Account:     n.Account,
		Region:      n.Region,
		Environment: n.Environment,
		Subnets:     subnets,
	}
	body, err := json.Marshal(webhook)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.auth))

	resp, err := p.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error creating network: %s", resp.Status)
	}

	pwr := &types.ProviderWebhookResponse{}
	err = json.NewDecoder(resp.Body).Decode(pwr)
	return pwr, err
}
