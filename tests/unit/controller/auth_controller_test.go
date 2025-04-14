package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"path/filepath"
	"testing"

	appctx "github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/controller"
	"github.com/not-empty/grit/app/util/jwt_manager"
	"github.com/stretchr/testify/require"
)

type fakeJwtManager struct{}

func (f *fakeJwtManager) Generate(audience, subject string, custom map[string]interface{}) string {
	return "fakeToken"
}

func (f *fakeJwtManager) IsValid(token string) (bool, error) {
	return true, nil
}

func (f *fakeJwtManager) IsOnTime(token string) (bool, error) {
	return true, nil
}

func (f *fakeJwtManager) TokenNeedsRefresh(token string) (bool, error) {
	return false, nil
}

func (f *fakeJwtManager) DecodePayload(token string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func setupTokenConfigFile(t *testing.T, content string) func() {
	err := os.MkdirAll("config", 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile("config/tokens.json", []byte(content), 0644)
	require.NoError(t, err)

	return func() {
		_ = os.Remove("config/tokens.json")
	}
}

func TestGenerate_InvalidJSON(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	ctrl := controller.NewAuthController()

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBufferString("invalid json"))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	bodyBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "Invalid JSON")
}

func TestGenerate_InvalidCredentials(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	ctrl := controller.NewAuthController()

	reqBody, err := json.Marshal(map[string]string{
		"token":  "api",
		"secret": "wrongsecret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	bodyBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "Invalid credentials")
}

func TestGenerate_Success(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController()

	reqBody, err := json.Marshal(map[string]string{
		"token":  "api",
		"secret": "testsecret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBuffer(reqBody))
	ctx := context.WithValue(req.Context(), appctx.RequestIDKey, "req-123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)

	token := rr.Header().Get("X-Token")
	expires := rr.Header().Get("X-Expires")
	reqID := rr.Header().Get("X-Request-ID")

	require.NotEmpty(t, token, "X-Token header should be set")
	require.NotEmpty(t, expires, "X-Expires header should be set")
	require.Equal(t, "req-123", reqID, "X-Request-ID header should match")

	_, err = time.Parse("2006-01-02 15:04:05", expires)
	require.NoError(t, err)
}

func TestLoadTokenConfig_AlreadyPopulated(t *testing.T) {
	prepopulated := map[string]controller.TokenConfig{
		"dummy": {Secret: "dummySecret", Name: "DummyApp"},
	}
	ac := &controller.AuthController{
		Config:     prepopulated,
		ConfigPath: "unused",
	}

	ac.LoadTokenConfig()
	require.Equal(t, prepopulated, ac.Config, "Config should remain unchanged if already populated")
}

func TestLoadTokenConfig_DecodeError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_controller_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "config")
	err = os.Mkdir(configDir, 0755)
	require.NoError(t, err)

	configFilePath := filepath.Join(configDir, "tokens.json")
	err = os.WriteFile(configFilePath, []byte("invalid json content"), 0644)
	require.NoError(t, err)

	ac := &controller.AuthController{
		Config:     make(map[string]controller.TokenConfig),
		ConfigPath: configFilePath,
	}

	var panicMessage string
	func() {
		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(string); ok {
					panicMessage = msg
				}
			}
		}()
		ac.LoadTokenConfig()
	}()

	require.NotEmpty(t, panicMessage, "Expected panic due to JSON decode error")
	require.Contains(t, panicMessage, "Error decoding tokens config:", "Panic message should indicate JSON decode error")
}

func TestLoadTokenConfig_DefaultPath(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "defaultSecret", "name": "DefaultApp"}}`)
	defer teardown()

	ac := &controller.AuthController{
		Config:     make(map[string]controller.TokenConfig),
		ConfigPath: "",
	}

	ac.LoadTokenConfig()

	cfg, ok := ac.Config["api"]
	require.True(t, ok, "Expected configuration key 'api' to be present")
	require.Equal(t, "defaultSecret", cfg.Secret, "Secret should be 'defaultSecret'")
	require.Equal(t, "DefaultApp", cfg.Name, "Name should be 'DefaultApp'")
}

func TestLoadTokenConfig_FileOpenError(t *testing.T) {
	ac := &controller.AuthController{
		Config:     make(map[string]controller.TokenConfig),
		ConfigPath: "nonexistingfile.json",
	}

	var panicMessage string
	func() {
		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(string); ok {
					panicMessage = msg
				} else {
					panicMessage = "panic not of type string"
				}
			}
		}()
		ac.LoadTokenConfig()
	}()

	require.NotEmpty(t, panicMessage, "Expected panic due to missing config file")
	require.Contains(t, panicMessage, "Error opening tokens config:", "Panic message should indicate file open error")
}

func TestGenerate_DefaultJwtSecret(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	os.Unsetenv("JWT_APP_SECRET")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController()

	reqBody, err := json.Marshal(map[string]string{
		"token":  "api",
		"secret": "testsecret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)

	token := rr.Header().Get("X-Token")
	require.NotEmpty(t, token, "Expected a token to be generated when JWT_APP_SECRET is not set")
}

func TestGenerate_DefaultExpire(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_EXPIRE", "0")
	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController()

	reqBody, err := json.Marshal(map[string]string{
		"token":  "api",
		"secret": "testsecret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)

	expHeader := rr.Header().Get("X-Expires")
	require.NotEmpty(t, expHeader, "X-Expires header should be set")

	expTime, err := time.Parse("2006-01-02 15:04:05", expHeader)
	require.NoError(t, err)

	now := time.Now()
	diff := expTime.Sub(now)

	require.True(t, diff.Seconds() >= 895 && diff.Seconds() <= 905,
		"expected expiration roughly 900 seconds from now, but got %.0f seconds", diff.Seconds())
}

func TestGenerate_DefaultRenew(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "0")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController()

	originalFactory := ctrl.JWTManagerFactory
	ctrl.JWTManagerFactory = func(secret, name string, expire, renew int64) jwt_manager.Manager {
		require.Equal(t, int64(300), renew, "Expected default renew value to be 300 when JWT_RENEW is 0")
		return &fakeJwtManager{}
	}
	defer func() {
		ctrl.JWTManagerFactory = originalFactory
	}()

	reqBody, err := json.Marshal(map[string]string{
		"token":  "api",
		"secret": "testsecret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)

	token := rr.Header().Get("X-Token")
	require.Equal(t, "fakeToken", token)
}

func TestGenerate_InternalError(t *testing.T) {
	teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "name": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController()

	ctrl.GenerateOverride = func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("test internal error")
	}

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBufferString(`{"token": "api", "secret": "testsecret"}`))
	req = req.WithContext(context.Background())
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test internal error", "Expected error message in response")
}
