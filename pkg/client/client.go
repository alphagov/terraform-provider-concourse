package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/concourse/concourse/go-concourse/concourse"
	"golang.org/x/oauth2"
)

// NewConcourseClient gives you an authenticated Concourse client using
// local user username and password authentication. Separate from Basic Auth.
func NewConcourseClient(
	url string,
	team string,
	username string,
	password string,
	caFile string,
	skipCertificateVerification bool,
) (concourse.Client, error) {

	oauth2Config := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",

		Endpoint: oauth2.Endpoint{TokenURL: url + "/sky/issuer/token"},
		Scopes:   []string{"email", "federated:id", "groups", "openid", "profile"},
	}
	cacerts, err := getCaCert(caFile)

	if err != nil {
		return nil, fmt.Errorf("Cannot load cacerts from file %s", caFile)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: cacerts, InsecureSkipVerify: skipCertificateVerification},
	}
	sslcli := &http.Client{Transport: tr}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	tok, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, err
	}
	tokenSource := oauth2.StaticTokenSource(tok)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	return concourse.NewClient(url, httpClient, true), nil
}

func getCaCert(caFileName string) (*x509.CertPool, error) {
	cacerts, err := ioutil.ReadFile(caFileName)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(cacerts)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	return roots, nil
}
