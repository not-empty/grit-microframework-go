package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	appctx "github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/util/jwt_manager"
)

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

		jwtSecret := os.Getenv("JWT_APP_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default_secret"
		}

		expireStr := os.Getenv("JWT_EXPIRE")
		renewStr := os.Getenv("JWT_RENEW")

		expire, err := strconv.ParseInt(expireStr, 10, 64)
		if err != nil || expireStr == "" {
			expire = 900
		}

		renew, err := strconv.ParseInt(renewStr, 10, 64)
		if err != nil || renewStr == "" {
			renew = 300
		}

		jwtMgr := jwt_manager.NewJwtManager(jwtSecret, contextHeader, expire, renew)

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
