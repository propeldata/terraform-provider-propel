package cms

import (
	"github.com/Khan/genqlient/graphql"
	"net/http"
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
	client := http.Client{}
	transport := client.Transport
	client.Transport = &withHeaders{
		headers:   headers,
		transport: transport,
	}
	return &client
}

func NewCmsClient(url string, oauthUrl string, clientId string, secret string) (*graphql.Client, error) {
	token, err := getToken(oauthUrl, clientId, secret)
	if err != nil {
		return nil, err
	}
	httpClient := newAuthenticatedHttpClientWithHeaders(map[string]string{"Authentication": "Bearer " + token})
	gqlClient := graphql.NewClient(url, httpClient)
	return &gqlClient, nil
}
