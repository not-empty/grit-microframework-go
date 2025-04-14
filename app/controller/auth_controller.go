package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/util/jwt_manager"

	appctx "github.com/not-empty/grit/app/context"
)

type TokenConfig struct {
	Secret string `json:"secret"`
	Name   string `json:"name"`
}

type AuthController struct {
	Config            map[string]TokenConfig
	ConfigPath        string
	JWTManagerFactory func(secret, name string, expire, renew int64) jwt_manager.Manager
	GenerateOverride  func(w http.ResponseWriter, r *http.Request) error
}

func NewAuthController() *AuthController {
	ac := &AuthController{
		Config:            make(map[string]TokenConfig),
		ConfigPath:        "config/tokens.json",
		JWTManagerFactory: jwt_manager.NewJwtManager,
	}
	ac.LoadTokenConfig()
	return ac
}

func (ac *AuthController) LoadTokenConfig() {
	if ac.Config != nil && len(ac.Config) > 0 {
		return
	}

	path := ac.ConfigPath
	if path == "" {
		path = "config/tokens.json"
	}

	file, err := os.Open(path)
	if err != nil {
		panic("Error opening tokens config: " + err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ac.Config); err != nil {
		panic("Error decoding tokens config: " + err.Error())
	}
}

func (ac *AuthController) Generate(w http.ResponseWriter, r *http.Request) {
	var err error
	if ac.GenerateOverride != nil {
		err = ac.GenerateOverride(w, r)
	} else {
		err = func() error {
			var req struct {
				Token  string `json:"token"`
				Secret string `json:"secret"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				helper.JSONError(w, http.StatusBadRequest, "Invalid JSON")
				return nil
			}

			cfg, ok := ac.Config[req.Token]
			if !ok || cfg.Secret != req.Secret {
				helper.JSONError(w, http.StatusUnauthorized, "Invalid credentials")
				return nil
			}

			jwtSecret := os.Getenv("JWT_APP_SECRET")
			if jwtSecret == "" {
				jwtSecret = "default_secret"
			}
			expire, _ := strconv.ParseInt(os.Getenv("JWT_EXPIRE"), 10, 64)
			if expire == 0 {
				expire = 900
			}
			renew, _ := strconv.ParseInt(os.Getenv("JWT_RENEW"), 10, 64)
			if renew == 0 {
				renew = 300
			}

			jwtMgr := ac.JWTManagerFactory(jwtSecret, cfg.Name, expire, renew)
			token := jwtMgr.Generate(cfg.Name, "api", map[string]interface{}{})
			expires := time.Now().Add(time.Duration(expire) * time.Second).Format("2006-01-02 15:04:05")

			w.Header().Set("X-Token", token)
			w.Header().Set("X-Expires", expires)
			if reqID, ok := r.Context().Value(appctx.RequestIDKey).(string); ok {
				w.Header().Set("X-Request-ID", reqID)
			}
			w.WriteHeader(http.StatusNoContent)
			return nil
		}()
	}
	if err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}
