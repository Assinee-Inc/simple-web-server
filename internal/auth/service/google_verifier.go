package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Assinee-Inc/docffy/internal/auth/model"
)

const defaultGoogleTokenInfoEndpoint = "https://oauth2.googleapis.com/tokeninfo"

var (
	ErrInvalidGoogleToken = errors.New("credencial Google inválida ou e-mail não verificado")
)

// GoogleTokenVerifier valida um ID Token OIDC emitido pelo Google.
type GoogleTokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*model.GoogleIdentity, error)
}

type googleOIDCVerifier struct {
	clientID   string
	endpoint   string
	httpClient *http.Client
}

type googleTokenInfoPayload struct {
	Iss           string `json:"iss"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Exp           any    `json:"exp"`
}

// NewGoogleOIDCVerifier cria um verificador padrão usando o endpoint tokeninfo oficial do Google.
func NewGoogleOIDCVerifier(clientID string) GoogleTokenVerifier {
	return NewGoogleOIDCVerifierWithEndpoint(clientID, defaultGoogleTokenInfoEndpoint, &http.Client{
		Timeout: 10 * time.Second,
	})
}

// NewGoogleOIDCVerifierWithEndpoint permite injetar endpoint e cliente HTTP customizados (ideal para testes).
func NewGoogleOIDCVerifierWithEndpoint(clientID, endpoint string, httpClient *http.Client) GoogleTokenVerifier {
	if endpoint == "" {
		endpoint = defaultGoogleTokenInfoEndpoint
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &googleOIDCVerifier{
		clientID:   strings.TrimSpace(clientID),
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

func (v *googleOIDCVerifier) VerifyIDToken(ctx context.Context, idToken string) (*model.GoogleIdentity, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, ErrInvalidGoogleToken
	}

	reqURL, err := url.Parse(v.endpoint)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}
	q := reqURL.Query()
	q.Set("id_token", idToken)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGoogleToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidGoogleToken
	}

	var payload googleTokenInfoPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidGoogleToken
	}

	if payload.Iss != "accounts.google.com" && payload.Iss != "https://accounts.google.com" {
		return nil, ErrInvalidGoogleToken
	}

	if v.clientID != "" && payload.Aud != v.clientID {
		return nil, ErrInvalidGoogleToken
	}

	expUnix, err := parseUnixTimestamp(payload.Exp)
	if err != nil || time.Now().UTC().Unix() >= expUnix {
		return nil, ErrInvalidGoogleToken
	}

	emailVerified := parseBoolClaim(payload.EmailVerified)
	if !emailVerified {
		return nil, ErrInvalidGoogleToken
	}

	if strings.TrimSpace(payload.Sub) == "" || strings.TrimSpace(payload.Email) == "" {
		return nil, ErrInvalidGoogleToken
	}

	displayName := strings.TrimSpace(payload.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(payload.GivenName + " " + payload.FamilyName)
	}
	if displayName == "" {
		parts := strings.Split(payload.Email, "@")
		displayName = parts[0]
	}

	return &model.GoogleIdentity{
		Sub:           strings.TrimSpace(payload.Sub),
		Email:         strings.TrimSpace(strings.ToLower(payload.Email)),
		EmailVerified: emailVerified,
		Name:          displayName,
	}, nil
}

func parseBoolClaim(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		return err == nil && b
	default:
		return false
	}
}

func parseUnixTimestamp(val any) (int64, error) {
	switch v := val.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	default:
		return 0, errors.New("exp inválido")
	}
}
