package client

import (
	"net/http"

	"github.com/Khan/genqlient/graphql"
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

func NewCmsClient(url string, oauthUrl string, clientId string, secret string) (graphql.Client, error) {
	token, err := getToken(oauthUrl, clientId, secret)
	if err != nil {
		return nil, err
	}
	httpClient := newAuthenticatedHttpClientWithHeaders(map[string]string{"Authorization": "Bearer " + token})

	gqlClient := graphql.NewClient(url, httpClient)
	return gqlClient, nil
}

//go:generate go run github.com/Khan/genqlient genqlient.yaml
