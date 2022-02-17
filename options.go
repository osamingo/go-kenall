package kenall

import "net/http"

type (
	withHTTPClient struct {
		client *http.Client
	}
	withEndpoint struct {
		endpoint string
	}
)

// Apply implements kenall.ClientOption interface.
func (w *withHTTPClient) Apply(cli *Client) {
	cli.HTTPClient = w.client
}

// Apply implements kenall.ClientOption interface.
func (w *withEndpoint) Apply(cli *Client) {
	cli.Endpoint = w.endpoint
}

// WithHTTPClient injects optional HTTP Client to kenall.Client.
func WithHTTPClient(cli *http.Client) ClientOption {
	return &withHTTPClient{client: cli}
}

// WithEndpoint injects optional endpoint to kenall.Client.
func WithEndpoint(endpoint string) ClientOption {
	return &withEndpoint{endpoint: endpoint}
}
