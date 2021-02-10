package kenall_test

import (
	"context"
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

	"github.com/osamingo/go-kenall"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		token string
		opts  []kenall.Option
		want  error
	}{
		"Empty token":         {token: "", opts: nil, want: kenall.ErrInvalidArgument},
		"Give token":          {token: "dummy", opts: nil, want: nil},
		"Give token and opts": {token: "dummy", opts: []kenall.Option{kenall.WithEndpoint(""), kenall.WithHTTPClient(nil)}, want: nil},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, c.opts...)
			if !errors.Is(c.want, err) {
				t.Errorf("give: %v, want: %v", err, c.want)
			}

			if len(c.opts) != 0 && cli.Endpoint != "" && cli.HTTPClient != nil {
				t.Error("option is not reflected")
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	t.Parallel()

	srv := runTestingServer(t)
	t.Cleanup(func() {
		srv.Close()
	})

	cases := map[string]struct {
		endpoint     string
		token        string
		postalcode   string
		checkAsError bool
		wantError    error
		wantJISX0402 string
	}{
		"Normal case":           {endpoint: srv.URL, token: "opencollector", postalcode: "1008105", checkAsError: false, wantError: nil, wantJISX0402: "13101"},
		"Invalid postalcode":    {endpoint: srv.URL, token: "opencollector", postalcode: "alphabet", checkAsError: false, wantError: kenall.ErrInvalidArgument, wantJISX0402: ""},
		"Not found":             {endpoint: srv.URL, token: "opencollector", postalcode: "0000000", checkAsError: false, wantError: kenall.ErrNotFound, wantJISX0402: ""},
		"Unauthorized":          {endpoint: srv.URL, token: "bad_token", postalcode: "0000000", checkAsError: false, wantError: kenall.ErrUnauthorized, wantJISX0402: ""},
		"Forbidden":             {endpoint: srv.URL, token: "opencollector", postalcode: "4030000", checkAsError: false, wantError: kenall.ErrForbidden, wantJISX0402: ""},
		"Internal server error": {endpoint: srv.URL, token: "opencollector", postalcode: "5000000", checkAsError: false, wantError: kenall.ErrInternalServerError, wantJISX0402: ""},
		"Bad gateway":           {endpoint: srv.URL, token: "opencollector", postalcode: "5020000", checkAsError: false, wantError: kenall.ErrBadGateway, wantJISX0402: ""},
		"Wrong endpoint":        {endpoint: "", token: "opencollector", postalcode: "5020000", checkAsError: true, wantError: &url.Error{}, wantJISX0402: ""},
		"Wrong response":        {endpoint: srv.URL, token: "opencollector", postalcode: "0000001", checkAsError: true, wantError: &json.MarshalerError{}, wantJISX0402: ""},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cli, err := kenall.NewClient(c.token, kenall.WithEndpoint(c.endpoint))
			if err != nil {
				t.Error(err)
			}

			res, err := cli.Get(context.Background(), c.postalcode)
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

func TestVersion_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		give      string
		want      time.Time
		wantError bool
	}{
		"Give 2020-11-30": {give: `"2020-11-30"`, want: time.Date(2020, 11, 30, 0, 0, 0, 0, time.UTC), wantError: false},
		"Give 20201130":   {give: `"20201130"`, want: time.Time{}, wantError: true},
		"Give null":       {give: `null`, want: time.Time{}, wantError: false},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := &kenall.Version{}
			err := v.UnmarshalJSON([]byte(c.give))
			if err == nil == c.wantError {
				t.Errorf("give: %v, want: %v", err, c.wantError)
			}
			if !c.want.Equal(time.Time(*v)) {
				t.Errorf("give: %v, want: %v", time.Time(*v), c.want)
			}
		})
	}
}

func ExampleClient_Get() {
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.Get(context.Background(), "1000001")
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

func runTestingServer(t *testing.T) *httptest.Server {
	t.Helper()

	const data = `{
  "version": "2020-11-30",
  "data": [
    {
      "jisx0402": "13101",
      "old_code": "100",
      "postal_code": "1008105",
      "prefecture_kana": "",
      "city_kana": "",
      "town_kana": "",
      "town_kana_raw": "",
      "prefecture": "東京都",
      "city": "千代田区",
      "town": "大手町",
      "koaza": "",
      "kyoto_street": "",
      "building": "",
      "floor": "",
      "town_partial": false,
      "town_addressed_koaza": false,
      "town_chome": false,
      "town_multi": false,
      "town_raw": "大手町",
      "corporation": {
        "name": "チッソ　株式会社",
        "name_kana": "ﾁﾂｿ ｶﾌﾞｼｷｶﾞｲｼﾔ",
        "block_lot": "２丁目２－１（新大手町ビル）",
        "post_office": "銀座",
        "code_type": 0
      }
    }
  ]
}`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.Fields(r.Header.Get("Authorization"))

		if len(token) != 2 || token[1] != "opencollector" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		switch r.URL.Path {
		case "/postalcode/1008105":
			if _, err := w.Write([]byte(data)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case "/postalcode/4030000":
			w.WriteHeader(http.StatusForbidden)
		case "/postalcode/5000000":
			w.WriteHeader(http.StatusInternalServerError)
		case "/postalcode/5020000":
			w.WriteHeader(http.StatusBadGateway)
		case "/postalcode/0000001":
			if _, err := w.Write([]byte("wrong")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
