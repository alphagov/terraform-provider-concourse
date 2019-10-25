package client

import (
	"net/http"
	"strings"
)

// AuthenticatedTransport is a transport which adds the Authorization header
type AuthenticatedTransport struct {
	AccessToken string
	TokenType   string
}

// RoundTrip represents a single authorized request/response cycle
func (t AuthenticatedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add(
		"Authorization",
		strings.Join([]string{t.TokenType, t.AccessToken}, " "),
	)

	return http.DefaultTransport.RoundTrip(r)
}
