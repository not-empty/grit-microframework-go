package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	appcontext "github.com/not-empty/grit/app/context"
	appmw "github.com/not-empty/grit/app/middleware"

	jwtmanager "github.com/not-empty/jwt-manager-go-lib"
	jwtmock "github.com/not-empty/jwt-manager-go-lib/mock"
)

func overrideJwt(mgr jwtmanager.Manager) func() {
	original := appmw.NewJwtManager
	appmw.NewJwtManager = func(secret, ctx string, expire, renew int64) jwtmanager.Manager {
		return mgr
	}
	return func() {
		appmw.NewJwtManager = original
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
	mgr := appmw.NewJwtManager("my-secret", "ctx", 100, 100)
	require.NotNil(t, mgr)
}

func TestJwtMiddleware_MissingHeaders(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	req := createReq("", "")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Missing authorization or context")
}

func TestJwtMiddleware_InvalidToken(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return false, errors.New("invalid")
		},
	}

	restore := overrideJwt(m)
	defer restore()

	req := createReq("bad-token", "ctx")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid or expired token")
}

func TestJwtMiddleware_ExpiredToken(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return false, errors.New("expired")
		},
	}

	restore := overrideJwt(m)
	defer restore()

	req := createReq("expired", "ctx")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Token expired or not yet valid")
}

func TestJwtMiddleware_BadPayload(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return true, nil
		},
		DecodePayloadFunc: func(token string) (map[string]interface{}, error) {
			return nil, errors.New("oops")
		},
	}

	restore := overrideJwt(m)
	defer restore()

	req := createReq("bad", "ctx")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Failed to decode token payload")
}

func TestJwtMiddleware_InvalidContext(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return true, nil
		},
		DecodePayloadFunc: func(token string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"aud": "wrong",
				"sub": "web",
				"exp": float64(time.Now().Unix() + 60),
			}, nil
		},
	}

	restore := overrideJwt(m)
	defer restore()

	req := createReq("mismatch", "ctx")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid token context or subject")
}

func TestJwtMiddleware_InvalidExpFormat(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return true, nil
		},
		DecodePayloadFunc: func(token string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"aud": "ctx",
				"sub": "api",
			}, nil
		},
	}

	restore := overrideJwt(m)
	defer restore()

	req := createReq("noexp", "ctx")
	rr := httptest.NewRecorder()

	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))
	handler.ServeHTTP(rr, req)

	expectUnauthorized(t, rr, "Invalid expiration format")
}

func TestJwtMiddleware_TokenRefresh(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return true, nil
		},
		DecodePayloadFunc: func(token string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"aud": "ctx",
				"sub": "api",
				"exp": float64(time.Now().Unix() + 60),
			}, nil
		},
		TokenNeedsRefreshFunc: func(token string) (bool, error) {
			return true, nil
		},
		GenerateFunc: func(aud, sub string, custom map[string]interface{}) string {
			return "new-token"
		},
	}

	restore := overrideJwt(m)
	defer restore()

	called := false
	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		info := r.Context().Value(appcontext.JwtContextKey).(appcontext.JwtTokenInfo)
		require.Equal(t, "new-token", info.Token)
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

	m := &jwtmock.JwtManagerMock{
		IsValidFunc: func(token string) (bool, error) {
			return true, nil
		},
		IsOnTimeFunc: func(token string) (bool, error) {
			return true, nil
		},
		DecodePayloadFunc: func(token string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"aud": "ctx",
				"sub": "api",
				"exp": float64(time.Now().Unix() + 300),
			}, nil
		},
		TokenNeedsRefreshFunc: func(token string) (bool, error) {
			return false, nil
		},
	}

	restore := overrideJwt(m)
	defer restore()

	called := false
	handler := appmw.JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		info := r.Context().Value(appcontext.JwtContextKey).(appcontext.JwtTokenInfo)
		require.Equal(t, "ok", info.Token)
	}))

	req := createReq("ok", "ctx")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}
