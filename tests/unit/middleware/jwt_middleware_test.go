package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/middleware"
	"github.com/not-empty/grit/app/util/jwt_manager"
)

type MockJwtManager struct {
	mock.Mock
}

func (m *MockJwtManager) IsValid(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func (m *MockJwtManager) IsOnTime(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func (m *MockJwtManager) DecodePayload(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockJwtManager) TokenNeedsRefresh(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func (m *MockJwtManager) Generate(aud, sub string, custom map[string]interface{}) string {
	args := m.Called(aud, sub, custom)
	return args.String(0)
}

func overrideJwt(mgr jwt_manager.Manager) func() {
	original := middleware.NewJwtManager
	middleware.NewJwtManager = func(secret, ctx string, expire, renew int64) jwt_manager.Manager {
		return mgr
	}
	return func() {
		middleware.NewJwtManager = original
	}
}

func createReq(token, ctx string) *http.Request {
	req := httptest.NewRequest("GET", "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if ctx != "" {
		req.Header.Set("Context", ctx)
	}
	return req
}

func expectUnauthorized(t *testing.T, rr *httptest.ResponseRecorder, msg string) {
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var res map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &res)
	require.NoError(t, err)
	require.Equal(t, msg, res["error"])
}

func setupEnv() {
	os.Setenv("JWT_APP_SECRET", "")
	os.Setenv("JWT_EXPIRE", "")
	os.Setenv("JWT_RENEW", "")
}

func TestDefaultNewJwtManager_IsCovered(t *testing.T) {
	manager := middleware.NewJwtManager("my-secret", "ctx", 100, 100)
	require.NotNil(t, manager)
}

func TestJwtMiddleware_MissingHeaders(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Missing authorization or context")
}

func TestJwtMiddleware_InvalidToken(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "bad-token").Return(false, errors.New("invalid"))

	restore := overrideJwt(mockJwt)
	defer restore()

	req := createReq("bad-token", "ctx")
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid or expired token")
}

func TestJwtMiddleware_ExpiredToken(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "expired").Return(true, nil)
	mockJwt.On("IsOnTime", "expired").Return(false, errors.New("expired"))

	restore := overrideJwt(mockJwt)
	defer restore()

	req := createReq("expired", "ctx")
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Token expired or not yet valid")
}

func TestJwtMiddleware_BadPayload(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "bad").Return(true, nil)
	mockJwt.On("IsOnTime", "bad").Return(true, nil)
	mockJwt.On("DecodePayload", "bad").Return(map[string]interface{}(nil), errors.New("oops"))

	restore := overrideJwt(mockJwt)
	defer restore()

	req := createReq("bad", "ctx")
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Failed to decode token payload")
}

func TestJwtMiddleware_InvalidContext(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "mismatch").Return(true, nil)
	mockJwt.On("IsOnTime", "mismatch").Return(true, nil)
	mockJwt.On("DecodePayload", "mismatch").Return(map[string]interface{}{
		"aud": "wrong", "sub": "web", "exp": float64(time.Now().Unix() + 60),
	}, nil)

	restore := overrideJwt(mockJwt)
	defer restore()

	req := createReq("mismatch", "ctx")
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid token context or subject")
}

func TestJwtMiddleware_InvalidExpFormat(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "noexp").Return(true, nil)
	mockJwt.On("IsOnTime", "noexp").Return(true, nil)
	mockJwt.On("DecodePayload", "noexp").Return(map[string]interface{}{
		"aud": "ctx", "sub": "api",
	}, nil)

	restore := overrideJwt(mockJwt)
	defer restore()

	req := createReq("noexp", "ctx")
	rr := httptest.NewRecorder()

	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid expiration format")
}

func TestJwtMiddleware_TokenRefresh(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "refresh").Return(true, nil)
	mockJwt.On("IsOnTime", "refresh").Return(true, nil)
	mockJwt.On("DecodePayload", "refresh").Return(map[string]interface{}{
		"aud": "ctx", "sub": "api", "exp": float64(time.Now().Unix() + 60),
	}, nil)
	mockJwt.On("TokenNeedsRefresh", "refresh").Return(true, nil)
	mockJwt.On("Generate", "ctx", "api", mock.Anything).Return("new-token")

	restore := overrideJwt(mockJwt)
	defer restore()

	called := false
	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		token := r.Context().Value(context.JwtContextKey).(context.JwtTokenInfo)
		require.Equal(t, "new-token", token.Token)
	}))

	req := createReq("refresh", "ctx")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestJwtMiddleware_HappyPath(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	mockJwt := new(MockJwtManager)
	mockJwt.On("IsValid", "ok").Return(true, nil)
	mockJwt.On("IsOnTime", "ok").Return(true, nil)
	mockJwt.On("DecodePayload", "ok").Return(map[string]interface{}{
		"aud": "ctx", "sub": "api", "exp": float64(time.Now().Unix() + 300),
	}, nil)
	mockJwt.On("TokenNeedsRefresh", "ok").Return(false, nil)

	restore := overrideJwt(mockJwt)
	defer restore()

	called := false
	handler := middleware.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		token := r.Context().Value(context.JwtContextKey).(context.JwtTokenInfo)
		require.Equal(t, "ok", token.Token)
	}))

	req := createReq("ok", "ctx")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}
