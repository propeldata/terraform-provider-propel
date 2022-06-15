package client

import (
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

const (
	apiURL   = "https://api.us-east-2.propeldata.com/graphql"
	oauthURL = "https://auth.us-east-2.propeldata.com/oauth2/token"
)

type withHeaders struct {
	headers   map[string]string
	transport http.RoundTripper
}

func (wh *withHeaders) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range wh.headers {
		req.Header.Add(k, v)
	}
	return wh.transport.RoundTrip(req)
}

// NewAuthenticatedHttpClientWithHeaders returns a new, authenticated HTTP client if the user is authenticated;
// otherwise, it prompts the user to authenticate before exiting with exit code 1.
//
// Additionally, it allows including default headers.
func newAuthenticatedHttpClientWithHeaders(headers map[string]string) *http.Client {
	client := http.DefaultClient
	client.Transport = &withHeaders{
		headers:   headers,
		transport: http.DefaultTransport,
	}
	return client
}

func NewPropelClient(clientId string, secret string) (graphql.Client, error) {
	token, err := getToken(oauthURL, clientId, secret)
	if err != nil {
		return nil, err
	}

	httpClient := newAuthenticatedHttpClientWithHeaders(map[string]string{"Authorization": "Bearer " + token})
	gqlClient := graphql.NewClient(apiURL, httpClient)

	return gqlClient, nil
}

//go:generate go run github.com/Khan/genqlient genqlient.yaml
