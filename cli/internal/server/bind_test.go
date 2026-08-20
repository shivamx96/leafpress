package server

import (
	"net"
	"strconv"
	"testing"
)

// resolveHost mirrors the default applied in Start.
func resolveHost(opts Options) string {
	if opts.Host == "" {
		return DefaultHost
	}
	return opts.Host
}

func TestDefaultHostIsLoopback(t *testing.T) {
	if DefaultHost != "127.0.0.1" {
		t.Fatalf("DefaultHost = %q, want loopback", DefaultHost)
	}
	if got := resolveHost(Options{}); got != "127.0.0.1" {
		t.Errorf("an unset Host must default to loopback, got %q", got)
	}
	if got := resolveHost(Options{Host: "0.0.0.0"}); got != "0.0.0.0" {
		t.Errorf("an explicit Host must be honoured, got %q", got)
	}
}

// The dev server has no authentication. Binding the default address must not
// accept connections from a non-loopback interface.
func TestDefaultBindRejectsExternalInterface(t *testing.T) {
	listener, err := net.Listen("tcp", net.JoinHostPort(DefaultHost, "0"))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected address type %T", listener.Addr())
	}
	if !addr.IP.IsLoopback() {
		t.Fatalf("default bind listens on %s, which is not loopback", addr.IP)
	}

	external := externalIPv4(t)
	if external == "" {
		t.Skip("no non-loopback IPv4 interface available")
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(external, strconv.Itoa(addr.Port)))
	if err == nil {
		conn.Close()
		t.Errorf("loopback-bound server accepted a connection via %s", external)
	}
}

func externalIPv4(t *testing.T) string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			return ip4.String()
		}
	}
	return ""
}

func TestDisplayURL(t *testing.T) {
	tests := map[string]string{
		"127.0.0.1":   "http://localhost:3000/",
		"0.0.0.0":     "http://localhost:3000/",
		"::":          "http://localhost:3000/",
		"192.168.1.5": "http://192.168.1.5:3000/",
	}
	for host, want := range tests {
		if got := displayURL(host, 3000); got != want {
			t.Errorf("displayURL(%q) = %q, want %q", host, got, want)
		}
	}
}
