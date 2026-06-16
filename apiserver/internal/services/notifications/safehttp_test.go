package notifications

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestIsDisallowedIP(t *testing.T) {
	cases := []struct {
		name       string
		ip         string
		disallowed bool
	}{
		{"loopback v4", "127.0.0.1", true},
		{"loopback range", "127.5.6.7", true},
		{"loopback v6", "::1", true},
		{"unspecified v4", "0.0.0.0", true},
		{"unspecified v6", "::", true},
		{"link-local metadata", "169.254.169.254", true},
		{"link-local v6", "fe80::1", true},
		{"private 10", "10.0.0.5", true},
		{"private 172", "172.16.3.4", true},
		{"private 192", "192.168.1.1", true},
		{"unique local v6", "fc00::1", true},
		{"multicast v4", "224.0.0.1", true},
		{"public v4", "8.8.8.8", false},
		{"public v4 alt", "1.1.1.1", false},
		{"public v6", "2606:4700:4700::1111", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if got := isDisallowedIP(ip); got != tc.disallowed {
				t.Fatalf("isDisallowedIP(%s) = %v, want %v", tc.ip, got, tc.disallowed)
			}
		})
	}
}

func TestIsDisallowedIPNil(t *testing.T) {
	if !isDisallowedIP(nil) {
		t.Fatal("nil IP must be disallowed")
	}
}

func TestValidateOutboundURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://example.com/hook", false},
		{"http", "http://example.com/hook", false},
		{"uppercase scheme", "HTTPS://example.com", false},
		{"file scheme", "file:///etc/passwd", true},
		{"gopher scheme", "gopher://example.com", true},
		{"ftp scheme", "ftp://example.com", true},
		{"no scheme", "example.com/hook", true},
		{"empty", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateOutboundURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateOutboundURL(%q) err = %v, wantErr %v", tc.url, err, tc.wantErr)
			}
		})
	}
}

func TestValidateOutboundMethod(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		want    string
		wantErr bool
	}{
		{"post", "POST", "POST", false},
		{"put", "PUT", "PUT", false},
		{"lowercase post", "post", "POST", false},
		{"padded put", " put ", "PUT", false},
		{"get", "GET", "", true},
		{"delete", "DELETE", "", true},
		{"empty", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateOutboundMethod(tc.method)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateOutboundMethod(%q) err = %v, wantErr %v", tc.method, err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("validateOutboundMethod(%q) = %q, want %q", tc.method, got, tc.want)
			}
		})
	}
}

// TestSafeClientBlocksLoopback verifies the dialer Control callback rejects a
// loopback connection at connect time even when the request is otherwise valid.
func TestSafeClientBlocksLoopback(t *testing.T) {
	client := newSafeHTTPClient(2 * time.Second)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://127.0.0.1:1/hook", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected loopback connection to be blocked")
	}

	if !strings.Contains(err.Error(), errDestinationNotAllowed) {
		t.Fatalf("expected dialer rejection error, got: %v", err)
	}
}
