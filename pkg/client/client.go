package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	httpClient *http.Client

	o *ClientOptions
}

type ClientOptions struct {
	Endpoint string
	Client   *http.Client
}

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	contextKeyClient = contextKey("client")
)

func NewClient(o *ClientOptions) *Client {
	return &Client{
		httpClient: o.Client,
		o:          o,
	}
}

func WithNewClient(ctx context.Context, o *ClientOptions) context.Context {
	c := NewClient(o)
	return context.WithValue(ctx, contextKeyClient, c)
}

func ClientFromContext(ctx context.Context) (*Client, bool) {
	c, ok := ctx.Value(contextKeyClient).(*Client)
	return c, ok
}

func (c *Client) baseUrl(path string) string {
	if strings.HasSuffix(c.o.Endpoint, "/") {
		return fmt.Sprintf("%s%s", c.o.Endpoint, path)
	}
	return fmt.Sprintf("%s/%s", c.o.Endpoint, path)
}
