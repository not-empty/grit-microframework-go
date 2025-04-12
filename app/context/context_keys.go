package context

type JwtTokenInfo struct {
	Token   string
	Expires string
}

type contextKey string

const (
	JwtContextKey contextKey = "jwtTokenInfo"
	RequestIDKey  contextKey = "request_id"
	AppVersionKey contextKey = "appVersion"
)
