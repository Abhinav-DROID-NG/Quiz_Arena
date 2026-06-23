package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// GoogleConfig holds the OAuth2 configuration for Google sign-in.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GoogleProvider wraps an OAuth2 config for Google.
type GoogleProvider struct {
	cfg *oauth2.Config
}

// GoogleUserInfo contains the user profile returned by Google's userinfo endpoint.
type GoogleUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Verified  bool   `json:"verified_email"`
}

// NewGoogleProvider creates a GoogleProvider using the supplied credentials.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &GoogleProvider{cfg: cfg}
}

// AuthCodeURL returns the URL the user should be redirected to in order to
// authenticate with Google. state is a CSRF-protection nonce.
func (p *GoogleProvider) AuthCodeURL(state string) string {
	return p.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange swaps the authorisation code received from Google's redirect
// for an OAuth2 token.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange: %w", err)
	}
	return token, nil
}

// UserInfo calls Google's userinfo endpoint and returns the authenticated user's profile.
func (p *GoogleProvider) UserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := p.cfg.Client(ctx, token)

	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("google userinfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google userinfo read: %w", err)
	}

	var info GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("google userinfo decode: %w", err)
	}

	if info.Email == "" {
		return nil, fmt.Errorf("google userinfo: empty email returned")
	}

	return &info, nil
}
