package notifications

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

var allowedOutboundMethods = map[string]struct{}{
	http.MethodPost: {},
	http.MethodPut:  {},
}

// isDisallowedIP reports whether an IP must not be reached by an outbound
// server-side request. It blocks loopback, link-local (including cloud metadata
// at 169.254.169.254), private, unspecified and multicast ranges.
func isDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	if ip.IsLoopback() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsPrivate() {
		return true
	}

	return false
}

// validateOutboundURL enforces the scheme allow-list and returns the parsed URL.
func validateOutboundURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL")
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return nil, fmt.Errorf("unsupported URL scheme")
	}

	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("invalid URL")
	}

	return parsed, nil
}

// validateOutboundMethod normalizes and enforces the method allow-list.
func validateOutboundMethod(method string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if _, ok := allowedOutboundMethods[normalized]; !ok {
		return "", fmt.Errorf("unsupported HTTP method")
	}

	return normalized, nil
}

// newSafeHTTPClient builds an HTTP client whose dialer validates the resolved IP
// at connect time. Checking the address inside Control closes the DNS-rebinding
// window between resolution and connection.
func newSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout: timeout,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("destination not allowed")
			}

			ip := net.ParseIP(host)
			if isDisallowedIP(ip) {
				return fmt.Errorf("destination not allowed")
			}

			return nil
		},
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
		Proxy: nil,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
