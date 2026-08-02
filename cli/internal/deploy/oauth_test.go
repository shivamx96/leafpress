package deploy

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGitHubOAuthDeviceExchange(t *testing.T) {
	requests := 0
	oauth := &GitHubOAuth{
		clientID: "test-client",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.Header.Get("Accept") != "application/json" && !strings.Contains(req.Header.Get("Accept"), "github+json") {
				t.Errorf("Accept = %q", req.Header.Get("Accept"))
			}
			status := http.StatusOK
			body := ""
			switch req.URL.String() {
			case githubDeviceCodeURL:
				assertFormValue(t, req, "client_id", "test-client")
				assertFormValue(t, req, "scope", githubScopes)
				body = `{"device_code":"device","user_code":"ABCD","verification_uri":"https://example.test/verify","expires_in":900,"interval":5}`
			case githubTokenURL:
				assertFormValue(t, req, "device_code", "device")
				body = `{"access_token":"token","token_type":"bearer","scope":"repo"}`
			case githubUserURL:
				if got := req.Header.Get("Authorization"); got != "Bearer token" {
					t.Errorf("Authorization = %q", got)
				}
				body = `{"login":"gardener","name":"Garden Author"}`
			default:
				status = http.StatusNotFound
			}
			return &http.Response{
				StatusCode: status,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		})},
	}

	device, err := oauth.requestDeviceCode(context.Background())
	if err != nil || device.DeviceCode != "device" || device.UserCode != "ABCD" {
		t.Fatalf("device response = %+v, %v", device, err)
	}
	token, err := oauth.checkToken(context.Background(), device.DeviceCode)
	if err != nil || token.AccessToken != "token" {
		t.Fatalf("token response = %+v, %v", token, err)
	}
	user, err := oauth.getUser(context.Background(), token.AccessToken)
	if err != nil || user.Login != "gardener" {
		t.Fatalf("user response = %+v, %v", user, err)
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want 3", requests)
	}
}

func TestGitHubOAuthRejectsTokenHTTPError(t *testing.T) {
	oauth := &GitHubOAuth{
		clientID: "test-client",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"error":"upstream"}`)),
				Request:    req,
			}, nil
		})},
	}
	if _, err := oauth.checkToken(context.Background(), "device"); err == nil || !strings.Contains(err.Error(), "token request failed") {
		t.Fatalf("error = %v", err)
	}
}

func assertFormValue(t *testing.T, req *http.Request, key, want string) {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = io.NopCloser(strings.NewReader(string(body)))
	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatal(err)
	}
	if got := values.Get(key); got != want {
		t.Errorf("form %s = %q, want %q", key, got, want)
	}
}
