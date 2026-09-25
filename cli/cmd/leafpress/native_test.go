package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Exercise an installed executable, filesystem events, and HTTP responses on
// the runner's OS. Cross-compiling alone cannot validate these behaviors.
func TestNativeInstallBuildAndWatch(t *testing.T) {
	binDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	install := exec.CommandContext(ctx, "go", "install", "-mod=readonly", ".")
	install.Env = append(os.Environ(), "GOBIN="+binDir, "GOWORK=off")
	if output, err := install.CombinedOutput(); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	binary := filepath.Join(binDir, "leafpress")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	garden := filepath.Join(t.TempDir(), "garden with spaces")
	if err := os.Mkdir(garden, 0755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Dir = garden
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output)
		}
	}
	run("version")
	run("init")
	run("new", "notes/native-page")
	if err := os.WriteFile(filepath.Join(garden, "index.md"), []byte("# Native garden\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("build", "--strict")
	if _, err := os.Stat(filepath.Join(garden, "_site", "notes", "native-page", "index.html")); !os.IsNotExist(err) {
		t.Fatalf("new draft was published: %v", err)
	}
	note := filepath.Join(garden, "notes", "native-page.md")
	if err := os.WriteFile(note, []byte("# Native published page\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("build", "--strict")
	if _, err := os.Stat(filepath.Join(garden, "_site", "notes", "native-page", "index.html")); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	logPath := filepath.Join(t.TempDir(), "serve.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()
	serve := exec.CommandContext(ctx, binary, "serve", "--port", strconv.Itoa(port))
	serve.Dir, serve.Stdout, serve.Stderr = garden, logFile, logFile
	if err := serve.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = serve.Process.Kill()
		_ = serve.Wait()
		if t.Failed() {
			data, _ := os.ReadFile(logPath)
			t.Logf("serve log:\n%s", data)
		}
	})
	waitFor := func(description string, ready func() bool) {
		t.Helper()
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			if ready() {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatalf("timed out waiting for %s", description)
	}
	var baseURL string
	address := regexp.MustCompile(`Server running at (http://[^\s]+)`)
	waitFor("watcher and server startup", func() bool {
		data, _ := os.ReadFile(logPath)
		match := address.FindStringSubmatch(string(data))
		if len(match) != 2 {
			return false
		}
		baseURL = match[1]
		return true
	})
	client := &http.Client{Timeout: time.Second}
	hasPage := func(want string) bool {
		response, err := client.Get(baseURL + "notes/native-page/")
		if err != nil {
			return false
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		return err == nil && response.StatusCode == http.StatusOK && strings.Contains(string(body), want)
	}
	waitFor("published page", func() bool { return hasPage("Native published page") })
	if err := os.WriteFile(note, []byte("# Native watcher refreshed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	waitFor("filesystem edit to rebuild the served page", func() bool { return hasPage("Native watcher refreshed") })
}
