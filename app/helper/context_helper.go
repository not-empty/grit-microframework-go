package helper

import (
	"context"
	"net/http"
)

func GetContextWithoutCancel(r *http.Request) context.Context {
	return context.WithoutCancel(r.Context())
}
