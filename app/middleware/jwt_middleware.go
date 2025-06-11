package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	appctx "github.com/not-empty/grit-microframework-go/app/context"

	"github.com/not-empty/grit-microframework-go/app/config"

	"github.com/not-empty/jwt-manager-go-lib"
)

var NewJwtManager = func(secret, context string, expire, renew int64) jwt_manager.Manager {
	return jwt_manager.NewJwtManager(secret, context, expire, renew)
}

func JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		contextHeader := r.Header.Get("Context")
		if tokenHeader == "" || contextHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing authorization or context"})
			return
		}

		token := strings.TrimPrefix(tokenHeader, "Bearer ")

		jwtSecret := config.AppConfig.JwtAppSecret
		expire := config.AppConfig.JwtExpire
		renew := config.AppConfig.JwtRenew

		jwtMgr := NewJwtManager(jwtSecret, contextHeader, expire, renew)

		if valid, err := jwtMgr.IsValid(token); err != nil || !valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token"})
			return
		}

		if onTime, err := jwtMgr.IsOnTime(token); err != nil || !onTime {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Token expired or not yet valid"})
			return
		}

		payload, err := jwtMgr.DecodePayload(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode token payload"})
			return
		}

		aud, _ := payload["aud"].(string)
		sub, _ := payload["sub"].(string)
		if aud != contextHeader || sub != "api" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token context or subject"})
			return
		}

		expFloat, ok := payload["exp"].(float64)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid expiration format"})
			return
		}

		expires := time.Unix(int64(expFloat), 0).Format("2006-01-02 15:04:05")

		if needsRefresh, err := jwtMgr.TokenNeedsRefresh(token); err == nil && needsRefresh {
			token = jwtMgr.Generate(aud, sub, map[string]interface{}{})
			expires = time.Now().Add(time.Duration(expire) * time.Second).Format("2006-01-02 15:04:05")
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, appctx.JwtContextKey, appctx.JwtTokenInfo{
			Token:   token,
			Expires: expires,
		})
		ctx = context.WithValue(ctx, appctx.AppVersionKey, "v1.0.0")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
