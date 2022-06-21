package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/olxbr/network-api/pkg/types"
)

func (c *Client) ListNetworks(ctx context.Context) ([]*types.Network, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl("api/v1/networks"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)

	var list types.NetworkListResponse
	if err := d.Decode(&list); err != nil {
		return nil, err
	}

	return list.Items, nil
}
