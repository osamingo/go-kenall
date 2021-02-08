package kenall_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/osamingo/go-kenall"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		token string
		opts  []kenall.Option
		want  error
	}{
		"Give token":  {token: "dummy", opts: nil, want: nil},
		"Empty token": {token: "", opts: nil, want: kenall.ErrInvalidArgument},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			_, err := kenall.NewClient(c.token, c.opts...)
			if err != c.want {
				t.Errorf("give: %v, want: %v", err, c.want)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	t.Parallel()

	srv := runTestingServer(t)
	defer srv.Close()

	// TODO: write test cases
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

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := strings.Fields(r.Header.Get("Authorization"))
		if len(token) != 2 || token[1] != "good_token" {
			return
		}

		w.Write([]byte(data))
	})

	return httptest.NewServer(h)
}
