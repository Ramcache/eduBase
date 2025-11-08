package middleware

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

// JWTVerifier — проверяет подпись токена и помещает claims в контекст
func JWTVerifier(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Verifier(tokenAuth)
}

// Authenticator — требует наличие валидного токена
func Authenticator(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Authenticator(tokenAuth)
}

// Проверка конкретной роли
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			if claims["role"] != role {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"access denied"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
