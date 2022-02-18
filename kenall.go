package kenall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const (
	// Endpoint is an endpoint provided by the kenall service.
	Endpoint = "https://api.kenall.jp/v1"
	// RFC3339DateFormat is the RFC3339-Date format for Go.
	RFC3339DateFormat = "2006-01-02"
)

type (
	// A Client implements API requests to the kenall service.
	Client struct {
		HTTPClient *http.Client
		Endpoint   string

		token string
	}
	// A ClientOption provides a customize option for kenall.Client.
	ClientOption interface {
		Apply(*Client)
	}

	// A GetAddressResponse is a result from the kenall service of the API to get the address from the postal code.
	GetAddressResponse struct {
		Version   Version    `json:"version"`
		Addresses []*Address `json:"data"`
	}
	// A GetCityResponse is a result from the kenall service of the API to get the city from the prefecture code.
	GetCityResponse struct {
		Version Version `json:"version"`
		Cities  []*City `json:"data"`
	}
)

// NewClient creates kenall.Client with the authorization token provided by the kenall service.
func NewClient(token string, opts ...ClientOption) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidArgument
	}

	cli := &Client{
		HTTPClient: http.DefaultClient,
		Endpoint:   Endpoint,
		token:      token,
	}

	for _, opt := range opts {
		opt.Apply(cli)
	}

	return cli, nil
}

// GetAddress requests to the kenall service to get the address by postal code.
func (cli *Client) GetAddress(ctx context.Context, postalCode string) (*GetAddressResponse, error) {
	if _, err := strconv.Atoi(postalCode); err != nil || len(postalCode) != 7 {
		return nil, ErrInvalidArgument
	}

	const path = "/postalcode/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+path+postalCode, nil)
	if err != nil {
		return nil, fmt.Errorf("kenall: failed to generate http request: %w", err)
	}

	var res GetAddressResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("kenall: failed to send request for kenall service: %w", err)
	}

	return &res, nil
}

// GetCity requests to the kenall service to get the city by prefecture code.
func (cli *Client) GetCity(ctx context.Context, prefectureCode string) (*GetCityResponse, error) {
	if _, err := strconv.Atoi(prefectureCode); err != nil || len(prefectureCode) != 2 {
		return nil, ErrInvalidArgument
	}

	const path = "/cities/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+path+prefectureCode, nil)
	if err != nil {
		return nil, fmt.Errorf("kenall: failed to generate http request: %w", err)
	}

	var res GetCityResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("kenall: failed to send request for kenall service: %w", err)
	}

	return &res, nil
}

// nolint: cyclop,goerr113
func (cli *Client) sendRequest(req *http.Request, res interface{}) error {
	req.Header.Add("Authorization", "token "+cli.token)

	resp, err := cli.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
			return ErrTimeout(err)
		}

		return fmt.Errorf("kenall: failed to do http client with a request for kenall service: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
		_, _ = io.Copy(ioutil.Discard, resp.Body)
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
			return fmt.Errorf("kenall: failed to decode to response: %w", err)
		}
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusPaymentRequired:
		return ErrPaymentRequired
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case http.StatusInternalServerError:
		return ErrInternalServerError
	default:
		return fmt.Errorf("kenall: not registered in the error handling, http status code = %d", resp.StatusCode)
	}

	return nil
}
