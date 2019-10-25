package client

import (
	"context"
	"net/http"

	"github.com/concourse/concourse/go-concourse/concourse"
	"golang.org/x/oauth2"
)

// NewConcourseClient gives you an authenticated Concourse client using
// local user username and password authentication. Separate from Basic Auth.
func NewConcourseClient(
	url string,
	team string,
	username string, password string,
) (concourse.Client, error) {

	client := concourse.NewClient(url, &http.Client{}, false)

	o2cfg := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",

		Endpoint: oauth2.Endpoint{TokenURL: client.URL() + "/sky/token"},
		Scopes:   []string{"email", "federated:id", "groups", "openid", "profile"},
	}

	ctx := context.WithValue(
		context.Background(),
		oauth2.HTTPClient, client.HTTPClient(),
	)

	tok, err := o2cfg.PasswordCredentialsToken(ctx, username, password)

	if err != nil {
		return nil, err
	}

	return concourse.NewClient(
		url,
		&http.Client{
			Transport: AuthenticatedTransport{
				AccessToken: tok.AccessToken,
				TokenType:   tok.TokenType,
			},
		}, false), nil
}
