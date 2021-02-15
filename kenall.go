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
	// ErrForbidden is an error value that will be returned when the resource does not have access privileges.
	ErrForbidden = fmt.Errorf("kenall: 403 forbidden error")
	// ErrNotFound is an error value that will be returned when there is no resource to be retrieved.
	ErrNotFound = fmt.Errorf("kenall: 404 not found error")
	// ErrInternalServerError is an error value that will be returned when some error occurs in the kenall service.
	ErrInternalServerError = fmt.Errorf("kenall: 500 internal server error")
	// ErrBadGateway is an error value that will be returned when the kenall service is unavailable.
	ErrBadGateway = fmt.Errorf("kenall: 502 bad gateway error")
)

type (
	// A Client implements API requests to the kenall service.
	Client struct {
		HTTPClient *http.Client
		Endpoint   string

		token string
	}
	// A Option provides a Functional Option Pattern for kenall.Client.
	Option func(*Client)
	// A GetAddressResponse is a result from the kenall service of the API to get the address from the postal code.
	GetAddressResponse struct {
		Version   Version    `json:"version"`
		Addresses []*Address `json:"data"`
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
)

// NewClient creates kenall.Client with the authorization token provided by the kenall.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidArgument
	}

	cli := &Client{
		HTTPClient: http.DefaultClient,
		Endpoint:   Endpoint,
		token:      token,
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli, nil
}

// WithHTTPClient injects optional HTTP Client to kenall.Client.
func WithHTTPClient(cli *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = cli
	}
}

// WithEndpoint injects optional endpoint to kenall.Client.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.Endpoint = endpoint
	}
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

	req.Header.Add("Authorization", "token "+cli.token)

	res, err := cli.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kenall: failed to request for kenall: %w", err)
	}
	defer func() {
		defer res.Body.Close()
		io.Copy(ioutil.Discard, res.Body)
	}()

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusInternalServerError:
		return nil, ErrInternalServerError
	case http.StatusBadGateway:
		return nil, ErrBadGateway
	}

	var resp GetAddressResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("kenall: failed to decode to response: %w", err)
	}

	return &resp, nil
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
