package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/olxbr/network-api/pkg/types"
)

func (c *Client) Version(ctx context.Context) (*types.Version, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
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

	var v *types.Version
	if err := d.Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
}
