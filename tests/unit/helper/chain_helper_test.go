package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit/app/helper"
)

func TestChain_NoMiddleware(t *testing.T) {
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Chain with no middleware
	helper.Chain(finalHandler).ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.String() != "final" {
		t.Errorf("Expected status 200 and body 'final', got %d and '%s'", w.Code, w.Body.String())
	}
}

func TestChain_WithMiddleware(t *testing.T) {
	wasCalled := false

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wasCalled = true
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("done"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	helper.Chain(finalHandler, middleware).ServeHTTP(w, req)

	if !wasCalled {
		t.Error("Expected middleware to be called")
	}
	if w.Body.String() != "done" {
		t.Errorf("Expected body 'done', got '%s'", w.Body.String())
	}
}
