package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrMissingToken     = errors.New("missing authorization token")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// SupabaseValidator validates Supabase JWT tokens using JWKS
type SupabaseValidator struct {
	jwksURL   string
	issuer    string
	audience  string
	keys      map[string]*rsa.PublicKey
	keysMutex sync.RWMutex
	client    *http.Client
}

// SupabaseClaims represents the claims in a Supabase JWT
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email         string                 `json:"email"`
	Phone         string                 `json:"phone"`
	Role          string                 `json:"role"`
	AppMetadata   map[string]interface{} `json:"app_metadata"`
	UserMetadata  map[string]interface{} `json:"user_metadata"`
	SessionID     string                 `json:"session_id"`
	AAL           string                 `json:"aal"`
	AMR           []AMREntry             `json:"amr"`
}

// AMREntry represents an authentication method reference
type AMREntry struct {
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewSupabaseValidator creates a new Supabase JWT validator
func NewSupabaseValidator(supabaseURL string) *SupabaseValidator {
	return &SupabaseValidator{
		jwksURL:  fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", strings.TrimSuffix(supabaseURL, "/")),
		issuer:   fmt.Sprintf("%s/auth/v1", strings.TrimSuffix(supabaseURL, "/")),
		audience: "authenticated",
		keys:     make(map[string]*rsa.PublicKey),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken validates a Supabase JWT and returns the user ID
func (v *SupabaseValidator) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, *SupabaseClaims, error) {
	// Parse the token without verification first to get the kid
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &SupabaseClaims{})
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Get the key ID from the token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return uuid.Nil, nil, fmt.Errorf("%w: missing kid in token header", ErrInvalidToken)
	}

	// Get the public key for this kid
	publicKey, err := v.getPublicKey(ctx, kid)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}

	// Parse and validate the token with the public key
	claims := &SupabaseClaims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, nil, ErrTokenExpired
		}
		return uuid.Nil, nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return uuid.Nil, nil, ErrInvalidToken
	}

	// Parse user ID from subject
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("%w: invalid user ID in token", ErrInvalidToken)
	}

	return userID, claims, nil
}

// getPublicKey retrieves the public key for the given kid
func (v *SupabaseValidator) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// Check cache first
	v.keysMutex.RLock()
	key, ok := v.keys[kid]
	v.keysMutex.RUnlock()
	if ok {
		return key, nil
	}

	// Fetch JWKS
	if err := v.fetchJWKS(ctx); err != nil {
		return nil, err
	}

	// Check cache again after fetch
	v.keysMutex.RLock()
	key, ok = v.keys[kid]
	v.keysMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

// fetchJWKS fetches the JWKS from Supabase
func (v *SupabaseValidator) fetchJWKS(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	// Parse and cache the keys
	v.keysMutex.Lock()
	defer v.keysMutex.Unlock()

	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Use != "sig" {
			continue
		}

		key, err := parseRSAPublicKey(jwk)
		if err != nil {
			continue
		}

		v.keys[jwk.Kid] = key
	}

	return nil
}

// parseRSAPublicKey parses an RSA public key from a JWK
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}
