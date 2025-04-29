package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/not-empty/grit/app/config"
	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/util/jwt_manager"

	appctx "github.com/not-empty/grit/app/context"
)

type TokenConfig struct {
	Secret  string `json:"secret"`
	Context string `json:"context"`
}

type AuthController struct {
	Config            map[string]TokenConfig
	ConfigPath        string
	JWTManagerFactory func(secret, context string, expire, renew int64) jwt_manager.Manager
	GenerateOverride  func(w http.ResponseWriter, r *http.Request) error
}

func NewAuthController(configPath string) *AuthController {
	if configPath == "" {
		configPath = "config/tokens.json"
	}

	ac := &AuthController{
		Config:            make(map[string]TokenConfig),
		ConfigPath:        configPath,
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
	if !filepath.IsAbs(path) {
		root, err := helper.GetProjectRoot()
		if err != nil {
			panic("Could not find project root: " + err.Error())
		}
		path = filepath.Join(root, path)
	}

	file, err := os.Open(path)
	if err != nil {
		panic("Could not open tokens config: " + err.Error())
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&ac.Config)
	if err != nil {
		panic("Could not decode tokens config: " + err.Error())
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
				helper.JSONError(w, http.StatusBadRequest, "Invalid JSON", err)
				return nil
			}

			cfg, ok := ac.Config[req.Token]
			if !ok || cfg.Secret != req.Secret {
				helper.JSONErrorSimple(w, http.StatusUnauthorized, "Invalid credentials")
				return nil
			}

			jwtSecret := config.AppConfig.JwtAppSecret
			expire := config.AppConfig.JwtExpire
			renew := config.AppConfig.JwtRenew

			jwtMgr := ac.JWTManagerFactory(jwtSecret, cfg.Context, expire, renew)
			token := jwtMgr.Generate(cfg.Context, "api", map[string]interface{}{})
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
		helper.JSONError(w, http.StatusInternalServerError, "Auth error", err)
	}
}
