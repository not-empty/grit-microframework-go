package controller

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

	"github.com/not-empty/grit/app/config"
	"github.com/not-empty/grit/app/controller"
	"github.com/not-empty/grit/app/helper"
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

func setAppEnvToLocal() {
	os.Setenv("APP_ENV", "local")
	_ = config.LoadConfig()
}

func setupTokenConfigFile(t *testing.T, content string) (configPath string, teardown func()) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")

	err := os.Mkdir(configDir, 0755)
	require.NoError(t, err)

	configPath = filepath.Join(configDir, "tokens.json")
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	return configPath, func() {
		_ = os.Remove(configPath)
	}
}

func TestGenerate_InvalidJSON(t *testing.T) {
	setAppEnvToLocal()
	configPath, teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "context": "TestApp"}}`)
	defer teardown()

	ctrl := controller.NewAuthController(configPath)

	req := httptest.NewRequest("POST", "/auth/generate", bytes.NewBufferString("invalid json"))
	rr := httptest.NewRecorder()

	ctrl.Generate(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	bodyBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyBytes), "Invalid JSON")
}

func TestGenerate_InvalidCredentials(t *testing.T) {
	setAppEnvToLocal()
	configPath, teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "context": "TestApp"}}`)
	defer teardown()

	ctrl := controller.NewAuthController(configPath)

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
	setAppEnvToLocal()

	configPath, teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "context": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController(configPath)

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

	require.NotEmpty(t, token)
	require.NotEmpty(t, expires)
	require.Equal(t, "req-123", reqID)

	_, err = time.Parse("2006-01-02 15:04:05", expires)
	require.NoError(t, err)
}

func TestLoadTokenConfig_AlreadyPopulated(t *testing.T) {
	setAppEnvToLocal()
	prepopulated := map[string]controller.TokenConfig{
		"dummy": {Secret: "dummySecret", Context: "DummyApp"},
	}
	ac := &controller.AuthController{
		Config:     prepopulated,
		ConfigPath: "unused",
	}

	ac.LoadTokenConfig()
	require.Equal(t, prepopulated, ac.Config, "Config should remain unchanged if already populated")
}

func TestLoadTokenConfig_DecodeError(t *testing.T) {
	setAppEnvToLocal()

	configPath, teardown := setupTokenConfigFile(t, `invalid json content`)
	defer teardown()

	ac := &controller.AuthController{
		Config:            make(map[string]controller.TokenConfig),
		ConfigPath:        configPath,
		JWTManagerFactory: nil,
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
	require.Contains(t, panicMessage, "Could not decode tokens config")
}

func TestLoadTokenConfig_FileOpenError(t *testing.T) {
	setAppEnvToLocal()
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
	require.Contains(t, panicMessage, "Could not open tokens config", "Panic message should indicate file open error")
}

func TestNewAuthController_DefaultConfigPath(t *testing.T) {
	setAppEnvToLocal()

	ac := &controller.AuthController{}
	defer func() {
		_ = recover()
		require.Equal(t, "config/tokens.json", ac.ConfigPath)
	}()

	ac = controller.NewAuthController("")
}

func TestLoadTokenConfig_ProjectRootError_ShouldPanic(t *testing.T) {
	setAppEnvToLocal()

	originalGetProjectRoot := helper.GetProjectRoot
	helper.GetProjectRoot = func() (string, error) {
		return "", fmt.Errorf("simulated project root error")
	}
	defer func() { helper.GetProjectRoot = originalGetProjectRoot }()

	ac := &controller.AuthController{
		Config:     nil,
		ConfigPath: "",
	}

	var panicMessage string
	func() {
		defer func() {
			if r := recover(); r != nil {
				if msg, ok := r.(string); ok {
					panicMessage = msg
				} else {
					panicMessage = "panic not a string"
				}
			}
		}()

		ac.LoadTokenConfig()
	}()

	require.NotEmpty(t, panicMessage, "Expected panic due to GetProjectRoot error")
	require.Contains(t, panicMessage, "Could not find project root: simulated project root error", "Panic should mention project root error")
}

func TestGenerate_InternalError(t *testing.T) {
	setAppEnvToLocal()
	configPath, teardown := setupTokenConfigFile(t, `{"api": {"secret": "testsecret", "context": "TestApp"}}`)
	defer teardown()

	os.Setenv("JWT_APP_SECRET", "myjwtsecret")
	os.Setenv("JWT_EXPIRE", "3600")
	os.Setenv("JWT_RENEW", "1800")
	defer os.Unsetenv("JWT_APP_SECRET")
	defer os.Unsetenv("JWT_EXPIRE")
	defer os.Unsetenv("JWT_RENEW")

	ctrl := controller.NewAuthController(configPath)

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
