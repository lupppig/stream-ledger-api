package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lupppig/stream-ledger-api/utils"
)

type CtxKey string

const ContextKeyUserID CtxKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			resp := utils.BuildResponse(http.StatusUnauthorized, "missing auth header", nil, nil)
			resp.BadResponse(w)
			return
		}

		var tokenStr string
		fmt.Sscanf(auth, "Bearer %s", &tokenStr)
		if tokenStr == "" {
			resp := utils.BuildResponse(http.StatusUnauthorized, "invalid auth header", nil, nil)
			resp.BadResponse(w)
			return
		}

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			resp := utils.BuildResponse(http.StatusUnauthorized, "invalid token", nil, err.Error())
			resp.BadResponse(w)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
