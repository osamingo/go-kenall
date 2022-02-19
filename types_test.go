package kenall_test

import (
	"testing"
	"time"

	"github.com/osamingo/go-kenall/v2"
)

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

func TestNullString_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		give      string
		want      string
		wantError bool
		isValid   bool
	}{
		"Give string": {give: `"123"`, want: "123", wantError: false, isValid: true},
		"Give number": {give: `123`, want: "", wantError: true, isValid: true},
		"Give empty":  {give: `""`, want: "", wantError: false, isValid: true},
		"Give null":   {give: `null`, want: "", wantError: false, isValid: false},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ns := &kenall.NullString{}
			err := ns.UnmarshalJSON([]byte(c.give))
			if err == nil == c.wantError {
				t.Fatalf("give: %v, want: %v", err, c.wantError)
			}
			if !c.isValid && ns.Valid {
				t.Errorf("give: %v, want: %v", ns.Valid, c.isValid)
			} else if c.want != ns.String {
				t.Errorf("give: %v, want: %v", ns.String, c.want)
			}
		})
	}
}

func TestRemoteAddress_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		give        string
		wantError   bool
		wantNetwork string
		wantAddress string
	}{
		"Give ip4":              {give: `{"type":"v4","address":"127.0.0.1"}`, wantError: false, wantNetwork: "ip", wantAddress: "127.0.0.1"},
		"Give ip6":              {give: `{"type":"v6","address":"::1"}`, wantError: false, wantNetwork: "ip", wantAddress: "::1"},
		"Give ip4 wrong object": {give: `{"type":"v4","address":"wrong"}`, wantError: true, wantNetwork: "", wantAddress: ""},
		"Give ip6 wrong object": {give: `{"type":"v6","address":"wrong"}`, wantError: true, wantNetwork: "", wantAddress: ""},
		"Give undefined type":   {give: `{"type":"v8","address":"::1"}`, wantError: true, wantNetwork: "", wantAddress: ""},
		"Give empty object":     {give: `{}`, wantError: true, wantNetwork: "", wantAddress: ""},
		"Give empty":            {give: ``, wantError: true, wantNetwork: "", wantAddress: ""},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ra := &kenall.RemoteAddress{}
			err := ra.UnmarshalJSON([]byte(c.give))
			if c.wantError {
				if err == nil {
					t.Errorf("an error should not be nil")
				}

				return
			}
			if err != nil {
				t.Fatalf("an error should be nil, err = %s", err)
			}
			if ra.Network() != c.wantNetwork {
				t.Errorf("give: %s, want: %s", ra.Network(), c.wantNetwork)
			}
			if ra.String() != c.wantAddress {
				t.Errorf("give: %s, want: %s", ra.String(), c.wantAddress)
			}
		})
	}
}
