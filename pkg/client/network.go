package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/olxbr/network-api/pkg/types"
)

func (c *Client) ListNetworks(ctx context.Context) (*types.NetworkListResponse, error) {
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
	if resp.StatusCode != http.StatusOK {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	ns := &types.NetworkListResponse{}
	if err := d.Decode(ns); err != nil {
		return nil, err
	}

	return ns, nil
}

func (c *Client) DetailNetwork(ctx context.Context, id string) (*types.Network, error) {
	url := c.baseUrl("api/v1/networks/" + id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	n := &types.Network{}
	if err := d.Decode(n); err != nil {
		return nil, err
	}

	return n, nil
}

func (c *Client) CreateNetwork(ctx context.Context, r *types.NetworkRequest) (*types.NetworkResponse, error) {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(r); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl("api/v1/networks"), buf)
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

	n := &types.NetworkResponse{}
	if err := d.Decode(n); err != nil {
		return nil, err
	}

	return n, nil
}
