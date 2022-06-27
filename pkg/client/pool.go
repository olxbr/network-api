package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/olxbr/network-api/pkg/types"
)

func (c *Client) ListPools(ctx context.Context) (*types.PoolResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl("api/v1/pools"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)

	p := &types.PoolResponse{}
	if err := d.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Client) CreatePool(ctx context.Context, r *types.PoolRequest) (*types.Pool, error) {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(r); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl("api/v1/pools"), buf)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	p := &types.Pool{}
	if err := d.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}
