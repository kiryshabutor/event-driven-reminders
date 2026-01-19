package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kiribu/jwt-practice/utils"
)

type contextKey string

const UserContextKey contextKey = "username"

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Отсутствует Authorization заголовок")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Неверный формат Authorization заголовка")
			return
		}

		tokenString := parts[1]

		claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Невалидный или истекший токен")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value(UserContextKey).(string)
	if !ok {
		return ""
	}
	return username
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
