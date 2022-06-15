package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CredentialsToken struct {
	AccessToken string `json:"access_token"`
	Expiry      string `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

//const URL = "https://auth.%s.propeldata.com/oauth2/token"

func getToken(oauthUrl string, clientId string, secret string) (string, error) {
	var token CredentialsToken

	payload := url.Values{}
	payload.Set("grant_type", "client_credentials")
	payload.Set("client_id", clientId)
	payload.Set("client_secret", secret)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, oauthUrl, strings.NewReader(payload.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_ = json.NewDecoder(resp.Body).Decode(&token)
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		bodyString := string(bodyBytes)

		return "", fmt.Errorf("Unable to generate Access Token (%d): %s\n\n", resp.StatusCode, bodyString)
	}
	return token.AccessToken, nil
}
