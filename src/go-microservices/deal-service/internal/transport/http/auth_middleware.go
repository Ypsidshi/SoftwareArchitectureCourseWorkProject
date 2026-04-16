package httptransport

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"coursework/platform-common/pkg/httpx"
	"github.com/golang-jwt/jwt/v5"
)

type authCtxKey string

const (
	clientIDCtxKey authCtxKey = "client_id"
	roleCtxKey     authCtxKey = "role"
)

func AuthMiddleware(jwtSecret string, logger *slog.Logger, allowedRoles ...string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing Authorization header"})
				return
			}
			if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid Authorization header"})
				return
			}
			tokenRaw := strings.TrimSpace(header[7:])
			token, err := jwt.Parse(tokenRaw, func(token *jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
			if err != nil || !token.Valid {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid token claims"})
				return
			}

			sub, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)
			if strings.TrimSpace(sub) == "" {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "token does not contain subject"})
				return
			}

			normalizedRole := strings.ToLower(strings.TrimSpace(role))
			if len(allowed) > 0 {
				if _, ok := allowed[normalizedRole]; !ok {
					logger.Warn("access denied for role", slog.String("role", role), slog.String("trace_id", httpx.TraceIDFromContext(r.Context())))
					httpx.WriteJSON(w, http.StatusForbidden, map[string]any{"error": "insufficient permissions"})
					return
				}
			}

			if normalizedRole == "" {
				httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "token does not contain role"})
				return
			}

			ctx := context.WithValue(r.Context(), clientIDCtxKey, sub)
			ctx = context.WithValue(ctx, roleCtxKey, normalizedRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClientAuthMiddleware(jwtSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return AuthMiddleware(jwtSecret, logger, "client")
}

func ClientIDFromContext(ctx context.Context) string {
	v, ok := ctx.Value(clientIDCtxKey).(string)
	if !ok {
		return ""
	}
	return v
}

func RoleFromContext(ctx context.Context) string {
	v, ok := ctx.Value(roleCtxKey).(string)
	if !ok {
		return ""
	}
	return v
}
