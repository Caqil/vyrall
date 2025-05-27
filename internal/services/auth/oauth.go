package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// OAuthService handles OAuth authentication
type OAuthService struct {
	configs map[string]*oauth2.Config
}

// OAuthUserInfo contains user information from OAuth providers
type OAuthUserInfo struct {
	ID       string
	Email    string
	Name     string
	Username string
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg *config.OAuthConfig) *OAuthService {
	configs := make(map[string]*oauth2.Config)

	// Set up Google OAuth
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		configs["google"] = &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes:       []string{"profile", "email"},
			Endpoint:     google.Endpoint,
		}
	}

	// Set up GitHub OAuth
	if cfg.GitHub.ClientID != "" && cfg.GitHub.ClientSecret != "" {
		configs["github"] = &oauth2.Config{
			ClientID:     cfg.GitHub.ClientID,
			ClientSecret: cfg.GitHub.ClientSecret,
			RedirectURL:  cfg.GitHub.RedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	// Set up Facebook OAuth
	if cfg.Facebook.ClientID != "" && cfg.Facebook.ClientSecret != "" {
		configs["facebook"] = &oauth2.Config{
			ClientID:     cfg.Facebook.ClientID,
			ClientSecret: cfg.Facebook.ClientSecret,
			RedirectURL:  cfg.Facebook.RedirectURL,
			Scopes:       []string{"email", "public_profile"},
			Endpoint:     facebook.Endpoint,
		}
	}

	return &OAuthService{
		configs: configs,
	}
}

// GetAuthURL returns the URL for OAuth authentication
func (s *OAuthService) GetAuthURL(provider, redirectURL string) (string, error) {
	provider = strings.ToLower(provider)
	config, ok := s.configs[provider]
	if !ok {
		return "", errors.New(errors.CodeInvalidArgument, fmt.Sprintf("Unsupported OAuth provider: %s", provider))
	}

	// Override redirect URL if provided
	if redirectURL != "" {
		config = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       config.Scopes,
			Endpoint:     config.Endpoint,
		}
	}

	// Generate state token
	state := generateRandomString(32)

	// Generate auth URL
	return config.AuthCodeURL(state), nil
}

// ExchangeCode exchanges an OAuth code for user information
func (s *OAuthService) ExchangeCode(ctx context.Context, provider, code string) (*OAuthUserInfo, error) {
	provider = strings.ToLower(provider)
	config, ok := s.configs[provider]
	if !ok {
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("Unsupported OAuth provider: %s", provider))
	}

	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to exchange OAuth code")
	}

	// Get user info based on provider
	switch provider {
	case "google":
		return s.getGoogleUserInfo(ctx, token)
	case "github":
		return s.getGitHubUserInfo(ctx, token)
	case "facebook":
		return s.getFacebookUserInfo(ctx, token)
	default:
		return nil, errors.New(errors.CodeInvalidArgument, "Unsupported OAuth provider")
	}
}

// GetSupportedProviders returns a list of supported OAuth providers
func (s *OAuthService) GetSupportedProviders() []string {
	providers := make([]string, 0, len(s.configs))
	for provider := range s.configs {
		providers = append(providers, provider)
	}
	return providers
}

// Helper methods for getting user info from different providers

func (s *OAuthService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Google user info")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read Google user info")
	}

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, errors.Wrap(err, "Failed to parse Google user info")
	}

	// Create username from email
	username := strings.Split(userInfo.Email, "@")[0]

	return &OAuthUserInfo{
		ID:       userInfo.Sub,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Username: username,
	}, nil
}

func (s *OAuthService) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get GitHub user info")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read GitHub user info")
	}

	var userInfo struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, errors.Wrap(err, "Failed to parse GitHub user info")
	}

	// If email is not public, get it from emails API
	if userInfo.Email == "" {
		emails, err := s.getGitHubEmails(ctx, client)
		if err != nil {
			return nil, err
		}
		if len(emails) > 0 {
			userInfo.Email = emails[0]
		}
	}

	return &OAuthUserInfo{
		ID:       fmt.Sprintf("%d", userInfo.ID),
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Username: userInfo.Login,
	}, nil
}

func (s *OAuthService) getGitHubEmails(ctx context.Context, client *http.Client) ([]string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get GitHub emails")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read GitHub emails")
	}

	var emailsInfo []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(data, &emailsInfo); err != nil {
		return nil, errors.Wrap(err, "Failed to parse GitHub emails")
	}

	// First try to find primary and verified email
	for _, email := range emailsInfo {
		if email.Primary && email.Verified {
			return []string{email.Email}, nil
		}
	}

	// Then try to find any verified email
	for _, email := range emailsInfo {
		if email.Verified {
			return []string{email.Email}, nil
		}
	}

	// Return all emails
	emails := make([]string, 0, len(emailsInfo))
	for _, email := range emailsInfo {
		emails = append(emails, email.Email)
	}

	return emails, nil
}

func (s *OAuthService) getFacebookUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email&access_token=" + token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Facebook user info")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read Facebook user info")
	}

	var userInfo struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, errors.Wrap(err, "Failed to parse Facebook user info")
	}

	// Create username from email or ID
	username := userInfo.ID
	if userInfo.Email != "" {
		username = strings.Split(userInfo.Email, "@")[0]
	}

	return &OAuthUserInfo{
		ID:       userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Username: username,
	}, nil
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	// Implementation using crypto/rand
	// This is a placeholder - replace with a proper implementation
	return "random_state_token"
}
