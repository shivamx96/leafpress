package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSiteImagePaths(t *testing.T) {
	for _, field := range []string{"favicon", "image"} {
		for _, path := range []string{"", "/static/uploads/icon.png", "/image.jpg", "/icons/a&b.png", "/icons/my%20icon.webp"} {
			c := Default()
			data, _ := json.Marshal(map[string]any{"site": map[string]string{field: path}})
			if err := json.Unmarshal(data, c); err != nil {
				t.Fatal(err)
			}
			if err := c.Validate(); err != nil {
				t.Errorf("%s=%q: %v", field, path, err)
			}
		}
		for _, path := range []string{"relative.png", "https://example.com/icon.png", "//example.com/icon.png", "/a/../icon.png", "/a/%2e%2e/icon.png", "/icon.png?q=1", "/icon.png?", "/icon.png#", "/a\\icon.png", "/%5cexample.com/icon.png", "/%2fexample.com/icon.png", "/%zz"} {
			c := Default()
			data, _ := json.Marshal(map[string]any{"site": map[string]string{field: path}})
			if err := json.Unmarshal(data, c); err != nil {
				t.Fatal(err)
			}
			if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "site."+field) {
				t.Errorf("%s=%q: want named validation error, got %v", field, path, err)
			}
		}
	}
}
