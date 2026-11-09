package config

import "testing"

func TestPublishConfig(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		invalid     bool
	}{
		{`{}`, "all", false},
		{`{"build":{"publish":"explicit"}}`, "explicit", false},
		{`{"build":{"publish":"typo"}}`, "", true},
	} {
		cfg, err := Parse([]byte(tc.input))
		if err == nil {
			err = cfg.Validate()
		}
		if tc.invalid {
			if err == nil {
				t.Fatal("accepted invalid publishing mode")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Build.Publish != tc.want {
			t.Fatalf("got %q want %q", cfg.Build.Publish, tc.want)
		}
	}
}
