package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/olxbr/network-api/pkg/types"
)

func (c *Client) ListProviders(ctx context.Context) (*types.ProviderListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl("api/v1/providers"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("error closing response body: %v", closeErr)
		}
	}()
	d := json.NewDecoder(resp.Body)

	p := &types.ProviderListResponse{}
	if err := d.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Client) CreateProvider(ctx context.Context, r *types.ProviderRequest) (*types.Provider, error) {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(r); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl("api/v1/providers"), buf)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("error closing response body: %v", closeErr)
		}
	}()

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	p := &types.Provider{}
	if err := d.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Client) UpdateProvider(ctx context.Context, name string, r *types.ProviderUpdateRequest) (*types.Provider, error) {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(r); err != nil {
		return nil, err
	}

	url := c.baseUrl("api/v1/providers/" + name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, buf)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("error closing response body: %v", closeErr)
		}
	}()

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	p := &types.Provider{}
	if err := d.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Client) DeleteProvider(ctx context.Context, name string) error {
	url := c.baseUrl("api/v1/providers/" + name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("error closing response body: %v", closeErr)
		}
	}()

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		e := &types.ErrorResponse{}
		if err := d.Decode(e); err != nil {
			return err
		}
		return fmt.Errorf("request failed %d: %+v", resp.StatusCode, e)
	}

	return nil
}
