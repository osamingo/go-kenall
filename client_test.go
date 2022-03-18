package kenall_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/osamingo/go-kenall/v2"
)

var (
	//go:embed testdata/addresses.json
	addressResponse []byte
	//go:embed testdata/cities.json
	cityResponse []byte
	//go:embed testdata/corporation.json
	corporationResponse []byte
	//go:embed testdata/whoami.json
	whoamiResponse []byte
	//go:embed testdata/holidays.json
	holidaysResponse []byte
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		token      string
		httpClient *http.Client
		endpoint   string
		want       error
	}{
		"Empty token":         {token: "", httpClient: nil, endpoint: "", want: kenall.ErrInvalidArgument},
		"Give token":          {token: "dummy", httpClient: nil, endpoint: "", want: nil},
		"Give token and opts": {token: "dummy", httpClient: &http.Client{}, endpoint: "customize_endpoint", want: nil},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := make([]kenall.ClientOption, 0, 2)
			if c.httpClient != nil {
				opts = append(opts, kenall.WithHTTPClient(c.httpClient))
			}
			if c.endpoint != "" {
				opts = append(opts, kenall.WithEndpoint(c.endpoint))
			}

			cli, err := kenall.NewClient(c.token, opts...)
			if !errors.Is(c.want, err) {
				t.Errorf("give: %v, want: %v", err, c.want)
			}

			if c.httpClient != nil && cli.HTTPClient != c.httpClient {
				t.Errorf("give: %v, want: %v", cli.HTTPClient, c.httpClient)
			}
			if c.endpoint != "" && cli.Endpoint != c.endpoint {
				t.Errorf("give: %v, want: %v", cli.Endpoint, c.endpoint)
			}
		})
	}
}

func TestClient_GetAddress(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint     string
		token        string
		ctx          context.Context
		postalCode   string
		checkAsError bool
		wantError    error
		wantJISX0402 string
	}{
		"Normal case":           {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "1008105", checkAsError: false, wantError: nil, wantJISX0402: "13104"},
		"Invalid postal code":   {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "alphabet", checkAsError: false, wantError: kenall.ErrInvalidArgument, wantJISX0402: ""},
		"Not found":             {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "0000000", checkAsError: false, wantError: kenall.ErrNotFound, wantJISX0402: ""},
		"Unauthorized":          {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), postalCode: "0000000", checkAsError: false, wantError: kenall.ErrUnauthorized, wantJISX0402: ""},
		"Payment Required":      {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "4020000", checkAsError: false, wantError: kenall.ErrPaymentRequired, wantJISX0402: ""},
		"Forbidden":             {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "4030000", checkAsError: false, wantError: kenall.ErrForbidden, wantJISX0402: ""},
		"Method Not Allowed":    {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "4050000", checkAsError: false, wantError: kenall.ErrMethodNotAllowed, wantJISX0402: ""},
		"Internal server error": {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "5000000", checkAsError: false, wantError: kenall.ErrInternalServerError, wantJISX0402: ""},
		"Unknown status code":   {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "5030000", checkAsError: true, wantError: fmt.Errorf(""), wantJISX0402: ""},
		"Wrong endpoint":        {endpoint: "", token: "opencollector", ctx: context.Background(), postalCode: "0000000", checkAsError: true, wantError: &url.Error{}, wantJISX0402: ""},
		"Wrong response":        {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), postalCode: "0000001", checkAsError: true, wantError: &json.MarshalerError{}, wantJISX0402: ""},
		"Nil context":           {endpoint: srv.URL, token: "opencollector", ctx: nil, postalCode: "0000000", checkAsError: true, wantError: errors.New("net/http: nil Context"), wantJISX0402: ""},
		"Timeout context":       {endpoint: srv.URL, token: "opencollector", ctx: toctx, postalCode: "1008105", checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantJISX0402: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetAddress(c.ctx, c.postalCode)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && res.Addresses[0].JISX0402 != c.wantJISX0402 {
				t.Errorf("give: %v, want: %v", res.Addresses[0].JISX0402, c.wantJISX0402)
			}
		})
	}
}

func TestClient_GetCity(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint       string
		token          string
		ctx            context.Context
		prefectureCode string
		checkAsError   bool
		wantError      error
		wantJISX0402   string
	}{
		"Normal case":             {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "13", checkAsError: false, wantError: nil, wantJISX0402: "13101"},
		"Invalid prefecture code": {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "alphabet", checkAsError: false, wantError: kenall.ErrInvalidArgument, wantJISX0402: ""},
		"Not found":               {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "48", checkAsError: false, wantError: kenall.ErrNotFound, wantJISX0402: ""},
		"Unauthorized":            {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), prefectureCode: "00", checkAsError: false, wantError: kenall.ErrUnauthorized, wantJISX0402: ""},
		"Payment Required":        {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "90", checkAsError: false, wantError: kenall.ErrPaymentRequired, wantJISX0402: ""},
		"Forbidden":               {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "91", checkAsError: false, wantError: kenall.ErrForbidden, wantJISX0402: ""},
		"Method Not Allowed":      {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "96", checkAsError: false, wantError: kenall.ErrMethodNotAllowed, wantJISX0402: ""},
		"Internal server error":   {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "92", checkAsError: false, wantError: kenall.ErrInternalServerError, wantJISX0402: ""},
		"Unknown status code":     {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "94", checkAsError: true, wantError: fmt.Errorf(""), wantJISX0402: ""},
		"Wrong endpoint":          {endpoint: "", token: "opencollector", ctx: context.Background(), prefectureCode: "00", checkAsError: true, wantError: &url.Error{}, wantJISX0402: ""},
		"Wrong response":          {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), prefectureCode: "95", checkAsError: true, wantError: &json.MarshalerError{}, wantJISX0402: ""},
		"Nil context":             {endpoint: srv.URL, token: "opencollector", ctx: nil, prefectureCode: "00", checkAsError: true, wantError: errors.New("net/http: nil Context"), wantJISX0402: ""},
		"Timeout context":         {endpoint: srv.URL, token: "opencollector", ctx: toctx, prefectureCode: "13", checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantJISX0402: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetCity(c.ctx, c.prefectureCode)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && res.Cities[0].JISX0402 != c.wantJISX0402 {
				t.Errorf("give: %v, want: %v", res.Cities[0].JISX0402, c.wantJISX0402)
			}
		})
	}
}

func TestClient_GetCorporation(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint        string
		token           string
		ctx             context.Context
		corporateNumber string
		checkAsError    bool
		wantError       error
		wantJISX0402    string
	}{
		"Normal case":              {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "2021001052596", checkAsError: false, wantError: nil, wantJISX0402: "13101"},
		"Invalid corporate number": {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "alphabet", checkAsError: false, wantError: kenall.ErrInvalidArgument, wantJISX0402: ""},
		"Not found":                {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000001", checkAsError: false, wantError: kenall.ErrNotFound, wantJISX0402: ""},
		"Unauthorized":             {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), corporateNumber: "2021001052596", checkAsError: false, wantError: kenall.ErrUnauthorized, wantJISX0402: ""},
		"Payment Required":         {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000402", checkAsError: false, wantError: kenall.ErrPaymentRequired, wantJISX0402: ""},
		"Forbidden":                {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000403", checkAsError: false, wantError: kenall.ErrForbidden, wantJISX0402: ""},
		"Method Not Allowed":       {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000405", checkAsError: false, wantError: kenall.ErrMethodNotAllowed, wantJISX0402: ""},
		"Internal server error":    {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000500", checkAsError: false, wantError: kenall.ErrInternalServerError, wantJISX0402: ""},
		"Unknown status code":      {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000503", checkAsError: true, wantError: fmt.Errorf(""), wantJISX0402: ""},
		"Wrong endpoint":           {endpoint: "", token: "opencollector", ctx: context.Background(), corporateNumber: "2021001052596", checkAsError: true, wantError: &url.Error{}, wantJISX0402: ""},
		"Wrong response":           {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), corporateNumber: "0000000000000", checkAsError: true, wantError: &json.MarshalerError{}, wantJISX0402: ""},
		"Nil context":              {endpoint: srv.URL, token: "opencollector", ctx: nil, corporateNumber: "2021001052596", checkAsError: true, wantError: errors.New("net/http: nil Context"), wantJISX0402: ""},
		"Timeout context":          {endpoint: srv.URL, token: "opencollector", ctx: toctx, corporateNumber: "2021001052596", checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantJISX0402: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetCorporation(c.ctx, c.corporateNumber)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && res.Corporation.JISX0402 != c.wantJISX0402 {
				t.Errorf("give: %v, want: %v", res.Corporation.JISX0402, c.wantJISX0402)
			}
		})
	}
}

func TestClient_GetWhoami(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint     string
		token        string
		ctx          context.Context
		checkAsError bool
		wantError    error
		wantAddr     string
	}{
		"Normal case":     {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), checkAsError: false, wantError: nil, wantAddr: "192.168.0.1"},
		"Unauthorized":    {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), checkAsError: false, wantError: kenall.ErrUnauthorized, wantAddr: ""},
		"Wrong endpoint":  {endpoint: "", token: "opencollector", ctx: context.Background(), checkAsError: true, wantError: &url.Error{}, wantAddr: ""},
		"Nil context":     {endpoint: srv.URL, token: "opencollector", ctx: nil, checkAsError: true, wantError: errors.New("net/http: nil Context"), wantAddr: ""},
		"Timeout context": {endpoint: srv.URL, token: "opencollector", ctx: toctx, checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantAddr: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetWhoami(c.ctx)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && res.RemoteAddress.String() != c.wantAddr {
				t.Errorf("give: %v, want: %v", res.RemoteAddress.String(), c.wantAddr)
			}
		})
	}
}

func TestClient_GetHolidays(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint     string
		token        string
		ctx          context.Context
		checkAsError bool
		wantError    error
		wantTitle    string
	}{
		"Normal case":     {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), checkAsError: false, wantError: nil, wantTitle: "元日"},
		"Unauthorized":    {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), checkAsError: false, wantError: kenall.ErrUnauthorized, wantTitle: ""},
		"Wrong endpoint":  {endpoint: "", token: "opencollector", ctx: context.Background(), checkAsError: true, wantError: &url.Error{}, wantTitle: ""},
		"Nil context":     {endpoint: srv.URL, token: "opencollector", ctx: nil, checkAsError: true, wantError: errors.New("net/http: nil Context"), wantTitle: ""},
		"Timeout context": {endpoint: srv.URL, token: "opencollector", ctx: toctx, checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantTitle: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetHolidays(c.ctx)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && res.Holidays[0].Title != c.wantTitle {
				t.Errorf("give: %v, want: %v", res.Holidays[0].Title, c.wantTitle)
			}
		})
	}
}

func TestClient_GetHolidaysByYear(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	cases := map[string]struct {
		endpoint     string
		token        string
		ctx          context.Context
		giveYear     int
		checkAsError bool
		wantError    error
		wantLen      int
	}{
		"Normal case":     {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), giveYear: 2022, checkAsError: false, wantError: nil, wantLen: 16},
		"Empty case":      {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), giveYear: 1969, checkAsError: false, wantError: nil, wantLen: 0},
		"Unauthorized":    {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), giveYear: 2022, checkAsError: false, wantError: kenall.ErrUnauthorized, wantLen: 0},
		"Wrong endpoint":  {endpoint: "", token: "opencollector", ctx: context.Background(), giveYear: 2022, checkAsError: true, wantError: &url.Error{}, wantLen: 0},
		"Nil context":     {endpoint: srv.URL, token: "opencollector", ctx: nil, giveYear: 2022, checkAsError: true, wantError: errors.New("net/http: nil Context"), wantLen: 0},
		"Timeout context": {endpoint: srv.URL, token: "opencollector", ctx: toctx, giveYear: 2022, checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantLen: 0},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetHolidaysByYear(c.ctx, c.giveYear)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && len(res.Holidays) != c.wantLen {
				t.Errorf("give: %v, want: %v", len(res.Holidays), c.wantLen)
			}
		})
	}
}

func TestClient_GetHolidaysByPeriod(t *testing.T) {
	t.Parallel()

	toctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	srv := runTestingServer(t)
	t.Cleanup(func() {
		cancel()
		srv.Close()
	})

	from, err := time.Parse(kenall.RFC3339DateFormat, "2022-01-01")
	if err != nil {
		t.Fatal(err)
	}

	to, err := time.Parse(kenall.RFC3339DateFormat, "2022-12-31")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		endpoint     string
		token        string
		ctx          context.Context
		giveFrom     time.Time
		giveTo       time.Time
		checkAsError bool
		wantError    error
		wantLen      int
	}{
		"Normal case":     {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), giveFrom: from, giveTo: to, checkAsError: false, wantError: nil, wantLen: 16},
		"Empty case":      {endpoint: srv.URL, token: "opencollector", ctx: context.Background(), giveFrom: from.Add(24 * time.Hour), giveTo: to, checkAsError: false, wantError: nil, wantLen: 0},
		"Unauthorized":    {endpoint: srv.URL, token: "bad_token", ctx: context.Background(), giveFrom: from, giveTo: to, checkAsError: false, wantError: kenall.ErrUnauthorized, wantLen: 0},
		"Wrong endpoint":  {endpoint: "", token: "opencollector", ctx: context.Background(), giveFrom: from, giveTo: to, checkAsError: true, wantError: &url.Error{}, wantLen: 0},
		"Nil context":     {endpoint: srv.URL, token: "opencollector", ctx: nil, giveFrom: from, giveTo: to, checkAsError: true, wantError: errors.New("net/http: nil Context"), wantLen: 0},
		"Timeout context": {endpoint: srv.URL, token: "opencollector", ctx: toctx, giveFrom: from, giveTo: to, checkAsError: true, wantError: kenall.ErrTimeout(context.DeadlineExceeded), wantLen: 0},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.GetHolidaysByPeriod(c.ctx, c.giveFrom, c.giveTo)
			if c.checkAsError && !errors.As(err, &c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			} else if !errors.Is(err, c.wantError) {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if res != nil && len(res.Holidays) != c.wantLen {
				t.Errorf("give: %v, want: %v", len(res.Holidays), c.wantLen)
			}
		})
	}
}

func ExampleClient_GetAddress() {
	if testing.Short() {
		// stab
		fmt.Print("false\n東京都 千代田区 千代田\n")

		return
	}

	// NOTE: Please set a valid token in the environment variable and run it.
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetAddress(context.Background(), "1000001")
	if err != nil {
		log.Fatal(err)
	}

	addr := res.Addresses[0]
	fmt.Println(time.Time(res.Version).IsZero())
	fmt.Println(addr.Prefecture, addr.City, addr.Town)
	// Output:
	// false
	// 東京都 千代田区 千代田
}

func ExampleClient_GetCity() {
	if testing.Short() {
		// stab
		fmt.Print("false\n東京都 千代田区\n")

		return
	}

	// NOTE: Please set a valid token in the environment variable and run it.
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetCity(context.Background(), "13")
	if err != nil {
		log.Fatal(err)
	}

	addr := res.Cities[0]
	fmt.Println(time.Time(res.Version).IsZero())
	fmt.Println(addr.Prefecture, addr.City)
	// Output:
	// false
	// 東京都 千代田区
}

func ExampleClient_GetCorporation() {
	if testing.Short() {
		// stab
		fmt.Print("false\n東京都 千代田区\n")

		return
	}

	// NOTE: Please set a valid token in the environment variable and run it.
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetCorporation(context.Background(), "7000012050002")
	if err != nil {
		log.Fatal(err)
	}

	corp := res.Corporation
	fmt.Println(time.Time(res.Version).IsZero())
	fmt.Println(corp.PrefectureName, corp.CityName)
	// Output:
	// false
	// 東京都 千代田区
}

func ExampleClient_GetWhoami() {
	if testing.Short() {
		// stab
		fmt.Println("ip")

		return
	}

	// NOTE: Please set a valid token in the environment variable and run it.
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetWhoami(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	raddr := res.RemoteAddress
	fmt.Println(raddr.IPAddr.Network())
	// Output:
	// ip
}

func ExampleClient_GetHolidays() {
	if testing.Short() {
		// stab
		fmt.Println("2022-01-01 元日")

		return
	}

	// NOTE: Please set a valid token in the environment variable and run it.
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetHolidaysByYear(context.Background(), 2022)
	if err != nil {
		log.Fatal(err)
	}

	day := res.Holidays[0]
	fmt.Println(day.Format(kenall.RFC3339DateFormat), day.Title)
	// Output:
	// 2022-01-01 元日
}

func runTestingServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.Fields(r.Header.Get("Authorization"))

		if len(token) != 2 || token[1] != "opencollector" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		switch uri := r.URL.RequestURI(); {
		case strings.HasPrefix(uri, "/postalcode/"):
			handlePostalAPI(t, w, uri)
		case strings.HasPrefix(uri, "/cities/"):
			handleCityAPI(t, w, uri)
		case strings.HasPrefix(uri, "/houjinbangou/"):
			handleCorporationAPI(t, w, uri)
		case strings.HasPrefix(uri, "/whoami"):
			handleWhoamiAPI(t, w, uri)
		case strings.HasPrefix(uri, "/holidays"):
			handleHolidaysAPI(t, w, uri)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func handlePostalAPI(t *testing.T, w http.ResponseWriter, uri string) {
	t.Helper()

	switch uri {
	case "/postalcode/1008105":
		if _, err := w.Write(addressResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case "/postalcode/4020000":
		w.WriteHeader(http.StatusPaymentRequired)
	case "/postalcode/4030000":
		w.WriteHeader(http.StatusForbidden)
	case "/postalcode/4050000":
		w.WriteHeader(http.StatusMethodNotAllowed)
	case "/postalcode/5000000":
		w.WriteHeader(http.StatusInternalServerError)
	case "/postalcode/5030000":
		w.WriteHeader(http.StatusServiceUnavailable)
	case "/postalcode/0000001":
		if _, err := w.Write([]byte("wrong")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleCityAPI(t *testing.T, w http.ResponseWriter, uri string) {
	t.Helper()

	switch uri {
	case "/cities/13":
		if _, err := w.Write(cityResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case "/cities/90":
		w.WriteHeader(http.StatusPaymentRequired)
	case "/cities/91":
		w.WriteHeader(http.StatusForbidden)
	case "/cities/92":
		w.WriteHeader(http.StatusInternalServerError)
	case "/cities/94":
		w.WriteHeader(http.StatusServiceUnavailable)
	case "/cities/95":
		if _, err := w.Write([]byte("wrong")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case "/cities/96":
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleCorporationAPI(t *testing.T, w http.ResponseWriter, uri string) {
	t.Helper()

	switch uri {
	case "/houjinbangou/2021001052596":
		if _, err := w.Write(corporationResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case "/houjinbangou/0000000000402":
		w.WriteHeader(http.StatusPaymentRequired)
	case "/houjinbangou/0000000000403":
		w.WriteHeader(http.StatusForbidden)
	case "/houjinbangou/0000000000405":
		w.WriteHeader(http.StatusMethodNotAllowed)
	case "/houjinbangou/0000000000500":
		w.WriteHeader(http.StatusInternalServerError)
	case "/houjinbangou/0000000000503":
		w.WriteHeader(http.StatusServiceUnavailable)
	case "/houjinbangou/0000000000000":
		if _, err := w.Write([]byte("wrong")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleWhoamiAPI(t *testing.T, w http.ResponseWriter, uri string) {
	t.Helper()

	switch uri {
	case "/whoami":
		if _, err := w.Write(whoamiResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleHolidaysAPI(t *testing.T, w http.ResponseWriter, uri string) {
	t.Helper()

	switch uri {
	case "/holidays?":
		if _, err := w.Write(holidaysResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	case "/holidays?year=2022":
		if _, err := w.Write(holidaysResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	case "/holidays?from=2022-01-01&to=2022-12-31":
		if _, err := w.Write(holidaysResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if strings.HasPrefix(uri, "/holidays") {
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNotFound)
}
