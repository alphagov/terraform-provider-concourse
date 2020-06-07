package client

import (
	"context"

	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/concourse/concourse/skymarshal/token"
	"golang.org/x/oauth2"
)

// NewConcourseClient gives you an authenticated Concourse client using
// local user username and password authentication. Separate from Basic Auth.
func NewConcourseClient(
	url string,
	team string,
	username string, password string,
) (concourse.Client, error) {

	oauth2Config := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",

		Endpoint: oauth2.Endpoint{TokenURL: url + "/sky/issuer/token"},
		Scopes:   []string{"email", "federated:id", "groups", "openid", "profile"},
	}

	ctx := context.Background()

	tok, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	tokenSource := oauth2.StaticTokenSource(tok)
	idTokenSource := token.NewTokenSource(tokenSource)
	httpClient := oauth2.NewClient(ctx, idTokenSource)

	return concourse.NewClient(url, httpClient, false), nil
}
