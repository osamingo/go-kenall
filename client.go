package kenall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Endpoint is an endpoint provided by the kenall service.
	Endpoint = "https://api.kenall.jp/v1"
	// RFC3339DateFormat is the RFC3339-Date format for Go.
	RFC3339DateFormat = "2006-01-02"

	errFailedGenerateRequestFormat = "kenall: failed to generate an http request: %w"
	errFailedRequestFormat         = "kenall: failed to send a request for kenall service: %w"
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

func (cli *Client) sendRequest(req *http.Request, res interface{}) error { //nolint: cyclop
	req.Header.Add("Authorization", "token "+cli.token)

	resp, err := cli.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
			return ErrTimeout(err)
		}

		return fmt.Errorf("kenall: failed to do http client with a request for kenall service: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
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
		//nolint: goerr113
		return fmt.Errorf("kenall: not registered in the error handling, http status code = %d", resp.StatusCode)
	}

	return nil
}

// A GetAddressResponse is a result from the kenall service of the API to get the address from the postal code.
type GetAddressResponse struct {
	Version   Version    `json:"version"`
	Addresses []*Address `json:"data"`
}

// GetAddress requests to the kenall service to get the address by postal code.
func (cli *Client) GetAddress(ctx context.Context, postalCode string) (*GetAddressResponse, error) {
	if _, err := strconv.Atoi(postalCode); err != nil || len(postalCode) != 7 {
		return nil, ErrInvalidArgument
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/postalcode/"+postalCode, nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetAddressResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// A GetCityResponse is a result from the kenall service of the API to get the city from the prefecture code.
type GetCityResponse struct {
	Version Version `json:"version"`
	Cities  []*City `json:"data"`
}

// GetCity requests to the kenall service to get the city by prefecture code.
func (cli *Client) GetCity(ctx context.Context, prefectureCode string) (*GetCityResponse, error) {
	if _, err := strconv.Atoi(prefectureCode); err != nil || len(prefectureCode) != 2 {
		return nil, ErrInvalidArgument
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/cities/"+prefectureCode, nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetCityResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// A GetCorporationResponse is a result from the kenall service of the API to get the corporation
// from the corporate number.
type GetCorporationResponse struct {
	Version     Version      `json:"version"`
	Corporation *Corporation `json:"data"`
}

// GetCorporation requests to the kenall service to get the corporation by corporate number.
func (cli *Client) GetCorporation(ctx context.Context, corporateNumber string) (*GetCorporationResponse, error) {
	if _, err := strconv.Atoi(corporateNumber); err != nil || len(corporateNumber) != 13 {
		return nil, ErrInvalidArgument
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/houjinbangou/"+corporateNumber, nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetCorporationResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// A GetWhoamiResponse is a result from the kenall service of the API to get whoami information.
type GetWhoamiResponse struct {
	RemoteAddress *RemoteAddress `json:"remote_addr"`
}

// GetWhoami requests to the kenall service to get the whoami information by access point.
func (cli *Client) GetWhoami(ctx context.Context) (*GetWhoamiResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/whoami", nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetWhoamiResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// A GetHolidaysResponse is a result from the kenall service of the API to get the holidays.
type GetHolidaysResponse struct {
	Holidays []*Holiday `json:"data"`
}

func (cli *Client) getHolidays(ctx context.Context, v url.Values) (*GetHolidaysResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/holidays?"+v.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetHolidaysResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// GetHolidays requests to the kenall service to get all holidays after 1970.
func (cli *Client) GetHolidays(ctx context.Context) (*GetHolidaysResponse, error) {
	return cli.getHolidays(ctx, nil)
}

// GetHolidaysByYear requests to the kenall service to get holidays for the year.
func (cli *Client) GetHolidaysByYear(ctx context.Context, year int) (*GetHolidaysResponse, error) {
	return cli.getHolidays(ctx, url.Values{"year": []string{strconv.Itoa(year)}})
}

// GetHolidaysByPeriod requests to the kenall service to get holidays for the period.
func (cli *Client) GetHolidaysByPeriod(ctx context.Context, from, to time.Time) (*GetHolidaysResponse, error) {
	return cli.getHolidays(ctx, url.Values{
		"from": []string{from.Format(RFC3339DateFormat)},
		"to":   []string{to.Format(RFC3339DateFormat)},
	})
}

// A GetNormalizeAddressResponse is a result from the kenall service of the API to normalize address.
type GetNormalizeAddressResponse struct {
	Version Version `json:"version"`
	Query   Query   `json:"query"`
}

// GetNormalizeAddress requests to the kenall service to normalize address.
func (cli *Client) GetNormalizeAddress(ctx context.Context, address string) (*GetNormalizeAddressResponse, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, ErrInvalidArgument
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.Endpoint+"/postalcode/?t="+address, nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	var res GetNormalizeAddressResponse
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &res, nil
}

// A GetBusinessDaysResponse is a result from the kenall service of the API to get the business days.
type GetBusinessDaysResponse struct {
	BusinessDay *BusinessDay
}

// GetBusinessDays requests to the kenall service to get business days by a date.
func (cli *Client) GetBusinessDays(ctx context.Context, date time.Time) (*GetBusinessDaysResponse, error) {
	if date.IsZero() {
		return nil, ErrInvalidArgument
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		cli.Endpoint+"/businessdays/check?date="+date.Format(RFC3339DateFormat), nil)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenerateRequestFormat, err)
	}

	res := struct {
		Result bool `json:"result"`
	}{}
	if err := cli.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf(errFailedRequestFormat, err)
	}

	return &GetBusinessDaysResponse{
		BusinessDay: &BusinessDay{
			LegalHoliday: res.Result,
			Time:         date,
		},
	}, nil
}
