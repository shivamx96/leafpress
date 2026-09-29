package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/shivamx96/leafpress/cli/internal/build"
	"github.com/shivamx96/leafpress/core/config"
)

// Ctrl+C cancels the serve context. That is a requested stop, so Start must
// return nil rather than surfacing http.ErrServerClosed as a failure.
func TestStartReturnsNilWhenContextIsCancelled(t *testing.T) {
	t.Chdir(t.TempDir())

	listener, err := net.Listen("tcp", net.JoinHostPort(DefaultHost, "0"))
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cfg := config.Default()
	cfg.Build.Port = port
	s := New(cfg, build.New(cfg, build.Options{}), Options{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx) }()

	client := &http.Client{Timeout: time.Second}
	url := "http://" + net.JoinHostPort(DefaultHost, strconv.Itoa(port)) + "/"
	deadline := time.Now().Add(10 * time.Second)
	for {
		response, err := client.Get(url)
		if err == nil {
			response.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start after cancellation = %v, want nil", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Start did not return after cancellation")
	}
}
