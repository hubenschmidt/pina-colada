package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	TenantIDKey  contextKey = "tenant_id"
	Auth0SubKey  contextKey = "auth0_sub"
	EmailKey     contextKey = "email"
	ClaimsKey    contextKey = "claims"
)

var (
	jwksCache jwk.Set
	jwksMu    sync.RWMutex
)

// Claims represents JWT claims from Auth0
type Claims struct {
	jwt.RegisteredClaims
	Email    string `json:"email"`
	TenantID *int64 `json:"https://pinacolada.co/tenant_id"`
}

// AuthMiddleware validates JWT tokens from Auth0
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractToken(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		claims, err := validateToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		ctx = context.WithValue(ctx, Auth0SubKey, claims.Subject)
		ctx = context.WithValue(ctx, EmailKey, getEmail(claims))

		// Check for tenant ID in header or claims
		if tenantID := getTenantID(r, claims); tenantID != nil {
			ctx = context.WithValue(ctx, TenantIDKey, *tenantID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	return token, nil
}

func validateToken(tokenString string) (*Claims, error) {
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	auth0Audience := os.Getenv("AUTH0_AUDIENCE")

	if auth0Domain == "" || auth0Audience == "" {
		return nil, fmt.Errorf("AUTH0_DOMAIN or AUTH0_AUDIENCE not configured")
	}

	keySet, err := getJWKS(auth0Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key not found for kid: %s", kid)
		}

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return nil, fmt.Errorf("failed to get raw key: %w", err)
		}

		return rawKey, nil
	}, jwt.WithAudience(auth0Audience), jwt.WithIssuer(fmt.Sprintf("https://%s/", auth0Domain)))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func getJWKS(domain string) (jwk.Set, error) {
	jwksMu.RLock()
	if jwksCache != nil {
		defer jwksMu.RUnlock()
		return jwksCache, nil
	}
	jwksMu.RUnlock()

	jwksMu.Lock()
	defer jwksMu.Unlock()

	// Double-check after acquiring write lock
	if jwksCache != nil {
		return jwksCache, nil
	}

	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", domain)
	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, err
	}

	jwksCache = set
	return jwksCache, nil
}

func getEmail(claims *Claims) string {
	if claims.Email != "" {
		return claims.Email
	}
	return ""
}

func getTenantID(r *http.Request, claims *Claims) *int64 {
	// First check header
	if tenantHeader := r.Header.Get("X-Tenant-Id"); tenantHeader != "" {
		var id int64
		if _, err := fmt.Sscanf(tenantHeader, "%d", &id); err == nil {
			return &id
		}
	}

	// Then check claims
	if claims.TenantID != nil {
		return claims.TenantID
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// GetUserID extracts user_id from context
func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}

// GetTenantID extracts tenant_id from context
func GetTenantID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(TenantIDKey).(int64)
	return id, ok
}

// UserLoader interface for loading users - implemented by AuthService
type UserLoader interface {
	GetOrCreateUser(auth0Sub, email string) (userID int64, tenantID *int64, err error)
}

// UserLoaderMiddleware loads user from DB and sets UserID in context
func UserLoaderMiddleware(loader UserLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			auth0Sub, ok := ctx.Value(Auth0SubKey).(string)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			email, _ := ctx.Value(EmailKey).(string)

			userID, tenantID, err := loader.GetOrCreateUser(auth0Sub, email)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to load user")
				return
			}

			ctx = context.WithValue(ctx, UserIDKey, userID)

			// Set tenant from user if not already set from header/claims
			if _, hasTenant := ctx.Value(TenantIDKey).(int64); !hasTenant && tenantID != nil {
				ctx = context.WithValue(ctx, TenantIDKey, *tenantID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuth0Sub extracts auth0_sub from context
func GetAuth0Sub(ctx context.Context) (string, bool) {
	sub, ok := ctx.Value(Auth0SubKey).(string)
	return sub, ok
}

// GetEmail extracts email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}
