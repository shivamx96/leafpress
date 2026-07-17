package main

import "testing"

func TestSelectVersion(t *testing.T) {
	tests := []struct {
		name   string
		linked string
		module string
		want   string
	}{
		{"release archive linker version wins", "v1.2.3", "v9.9.9", "v1.2.3"},
		{"go install module version", "dev", "v1.2.3", "v1.2.3"},
		{"local build", "dev", "(devel)", "dev"},
		{"missing build info", "dev", "", "dev"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := selectVersion(test.linked, test.module); got != test.want {
				t.Fatalf("selectVersion(%q, %q) = %q, want %q", test.linked, test.module, got, test.want)
			}
		})
	}
}
