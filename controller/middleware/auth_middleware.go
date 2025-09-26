package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lupppig/stream-ledger-api/utils"
)

type ctxKey string

const contextKeyUserID ctxKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing auth header", http.StatusUnauthorized)
			return
		}
		var tokenStr string
		fmt.Sscanf(auth, "Bearer %s", &tokenStr)
		fmt.Println("token str", tokenStr)
		if tokenStr == "" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
