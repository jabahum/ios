package keycloak

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
}

type Client struct {
	InternalURL  string
	PublicURL    string
	Realm        string
	ClientID     string
	ClientSecret string
	httpClient   *http.Client
}

type UserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

func NewClient(
	internalURL string,
	publicURL string,
	realm string,
	clientID string,
	clientSecret string,
) (*Client, error) {

	if internalURL == "" {
		return nil, fmt.Errorf("keycloak internalURL is required")
	}
	if publicURL == "" {
		return nil, fmt.Errorf("keycloak publicURL is required")
	}
	if realm == "" {
		return nil, fmt.Errorf("keycloak realm is required")
	}
	if clientID == "" {
		return nil, fmt.Errorf("keycloak clientID is required")
	}

	// Normalize URLs
	internalURL = strings.TrimSpace(internalURL)
	internalURL = strings.TrimSuffix(internalURL, "/")

	publicURL = strings.TrimSpace(publicURL)
	publicURL = strings.TrimSuffix(publicURL, "/")

	// Sanity checks
	if !strings.HasPrefix(internalURL, "http://") && !strings.HasPrefix(internalURL, "https://") {
		return nil, fmt.Errorf(
			"keycloak internalURL must start with http:// or https:// (got %q)",
			internalURL,
		)
	}

	if !strings.HasPrefix(publicURL, "http://") && !strings.HasPrefix(publicURL, "https://") {
		return nil, fmt.Errorf(
			"keycloak publicURL must start with http:// or https:// (got %q)",
			publicURL,
		)
	}

	// Warn about localhost usage (expected only for public URL)
	if strings.Contains(internalURL, "localhost") {
		log.Printf(
			"[WARN][KC] internalURL contains 'localhost'. "+
				"This will NOT work across containers. internalURL=%s",
			internalURL,
		)
	}

	client := &Client{
		InternalURL:  internalURL,
		PublicURL:    publicURL,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	log.Printf(
		"[KC] Initialized client internalURL=%s publicURL=%s realm=%s clientID=%s",
		client.InternalURL,
		client.PublicURL,
		client.Realm,
		client.ClientID,
	)

	return client, nil
}

func (c *Client) ExchangeCodeForToken(
	code string,
	redirectURI string,
	codeVerifier string,
) (*TokenResponse, error) {

	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.InternalURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {

		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenRes, nil
}

func (c *Client) AccessToken(refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.InternalURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token failed [%d]: %s", res.StatusCode, string(bodyBytes))
	}

	var tokenRes TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenRes); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenRes, nil
}

func (c *Client) LogOut(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("missing refresh token for logout")
	}

	logoutURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/logout",
		c.InternalURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("refresh_token", refreshToken)

	res, err := c.httpClient.PostForm(logoutURL, form)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("logout failed [%d]: %s", res.StatusCode, string(body))
	}

	return nil
}

func (c *Client) UserInfo(accessToken string) (*UserInfo, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"%s/realms/%s/protocol/openid-connect/userinfo",
			c.InternalURL,
			c.Realm,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("userinfo failed [%d]: %s", res.StatusCode, body)
	}

	var ui UserInfo
	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return nil, err
	}

	return &ui, nil
}

func (c *Client) BuildLoginURL(
	redirectURI string,
	state string,
	codeChallenge string,
) string {

	q := url.Values{}
	q.Set("client_id", c.ClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")

	return fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/auth?%s",
		c.PublicURL,
		c.Realm,
		q.Encode(),
	)
}
