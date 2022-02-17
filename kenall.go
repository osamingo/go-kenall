package kenall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	// Endpoint is an endpoint provided by the kenall service.
	Endpoint = "https://api.kenall.jp/v1"
	// RFC3339DateFormat is the RFC3339-Date format for Go.
	RFC3339DateFormat = "2006-01-02"
)

var (
	// ErrInvalidArgument is an error value that will be returned if the value of the argument is invalid.
	ErrInvalidArgument = fmt.Errorf("kenall: invalid argument")
	// ErrUnauthorized is an error value that will be returned if the authorization token is invalid.
	ErrUnauthorized = fmt.Errorf("kenall: 401 unauthorized error")
	// ErrPaymentRequired is an error value that will be returned if the payment for your kenall account is overdue.
	ErrPaymentRequired = fmt.Errorf("kenall: 402 payment required error")
	// ErrForbidden is an error value that will be returned when the resource does not have access privileges.
	ErrForbidden = fmt.Errorf("kenall: 403 forbidden error")
	// ErrNotFound is an error value that will be returned when there is no resource to be retrieved.
	ErrNotFound = fmt.Errorf("kenall: 404 not found error")
	// ErrMethodNotAllowed is an error value that will be returned when the request calls a method that is not allowed.
	ErrMethodNotAllowed = fmt.Errorf("kenall: 405 method not allowed error")
	// ErrInternalServerError is an error value that will be returned when some error occurs in the kenall service.
	ErrInternalServerError = fmt.Errorf("kenall: 500 internal server error")
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
	withHTTPClient struct {
		client *http.Client
	}
	withEndpoint struct {
		endpoint string
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

	// A Version is the version-controlled date of the retrieved data.
	Version time.Time
	// An Address is an address associated with the postal code defined by JP POST.
	Address struct {
		JISX0402           string `json:"jisx0402"`
		OldCode            string `json:"old_code"`
		PostalCode         string `json:"postal_code"`
		PrefectureKana     string `json:"prefecture_kana"`
		CityKana           string `json:"city_kana"`
		TownKana           string `json:"town_kana"`
		TownKanaRaw        string `json:"town_kana_raw"`
		Prefecture         string `json:"prefecture"`
		City               string `json:"city"`
		Town               string `json:"town"`
		Koaza              string `json:"koaza"`
		KyotoStreet        string `json:"kyoto_street"`
		Building           string `json:"building"`
		Floor              string `json:"floor"`
		TownPartial        bool   `json:"town_partial"`
		TownAddressedKoaza bool   `json:"town_addressed_koaza"`
		TownChome          bool   `json:"town_chome"`
		TownMulti          bool   `json:"town_multi"`
		TownRaw            string `json:"town_raw"`
		Corporation        struct {
			Name       string `json:"name"`
			NameKana   string `json:"name_kana"`
			BlockLot   string `json:"block_lot"`
			PostOffice string `json:"post_office"`
			CodeType   int    `json:"code_type"`
		} `json:"corporation"`
	}
	// A City is a city associated with the prefecture code defined by JIS X 0401.
	City struct {
		JISX0402       string `json:"jisx0402"`
		PrefectureCode string `json:"prefecture_code"`
		CityCode       string `json:"city_code"`
		PrefectureKana string `json:"prefecture_kana"`
		CityKana       string `json:"city_kana"`
		Prefecture     string `json:"prefecture"`
		City           string `json:"city"`
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

// WithHTTPClient injects optional HTTP Client to kenall.Client.
func WithHTTPClient(cli *http.Client) ClientOption {
	return &withHTTPClient{client: cli}
}

// WithEndpoint injects optional endpoint to kenall.Client.
func WithEndpoint(endpoint string) ClientOption {
	return &withEndpoint{endpoint: endpoint}
}

// GetAddress requests to the kenall service to get the address by postal code.
func (cli *Client) GetAddress(ctx context.Context, postalCode string) (*GetAddressResponse, error) {
	if _, err := strconv.Atoi(postalCode); err != nil || len(postalCode) != 7 {
		return nil, ErrInvalidArgument
	}

	const path = "/postalCode/"

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
		return fmt.Errorf("kenall: failed to do http client with a request for kenall service: %w", err)
	}

	defer func() {
		defer resp.Body.Close()
		io.Copy(ioutil.Discard, resp.Body)
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// do nothing :)
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

	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return fmt.Errorf("kenall: failed to decode to response: %w", err)
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Version) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	t, err := time.Parse(`"`+RFC3339DateFormat+`"`, string(data))
	if err != nil {
		return fmt.Errorf("kenall: failed to parse date with RFC3339 Date: %w", err)
	}

	*v = Version(t)

	return nil
}

// Apply implements kenall.ClientOption interface.
func (w *withHTTPClient) Apply(cli *Client) {
	cli.HTTPClient = w.client
}

// Apply implements kenall.ClientOption interface.
func (w *withEndpoint) Apply(cli *Client) {
	cli.Endpoint = w.endpoint
}
