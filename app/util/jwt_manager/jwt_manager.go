package jwt_manager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Manager interface {
	Generate(audience, subject string, custom map[string]interface{}) string
	IsValid(token string) (bool, error)
	IsOnTime(token string) (bool, error)
	TokenNeedsRefresh(token string) (bool, error)
	DecodePayload(token string) (map[string]interface{}, error)
}

var _ Manager = (*JwtManager)(nil)

type JwtManager struct {
	AppSecret string
	Context   string
	Expire    int64
	Renew     int64
	algorithm string
	tokenType string
}

func NewJwtManager(secret, context string, expire, renew int64) Manager {
	return &JwtManager{
		AppSecret: secret,
		Context:   context,
		Expire:    expire,
		Renew:     renew,
		algorithm: "HS256",
		tokenType: "JWT",
	}
}

func (j *JwtManager) GetHeader() string {
	header := map[string]string{
		"alg": j.algorithm,
		"typ": j.tokenType,
	}
	data, _ := json.Marshal(header)
	return Base64UrlEncode(data)
}

func (j *JwtManager) getPayload(audience, subject string, custom map[string]interface{}) string {
	now := time.Now().Unix()
	payload := map[string]interface{}{
		"aud": audience,
		"exp": now + j.Expire,
		"iat": now,
		"iss": j.Context,
		"sub": subject,
	}
	for k, v := range custom {
		payload[k] = v
	}
	data, _ := json.Marshal(payload)
	return Base64UrlEncode(data)
}

func (j *JwtManager) GetSignature(header, payload string) string {
	h := hmac.New(sha256.New, []byte(j.AppSecret))
	h.Write([]byte(header + "." + payload))
	return Base64UrlEncode(h.Sum(nil))
}

func (j *JwtManager) Generate(audience, subject string, custom map[string]interface{}) string {
	header := j.GetHeader()
	payload := j.getPayload(audience, subject, custom)
	signature := j.GetSignature(header, payload)
	return header + "." + payload + "." + signature
}

func (j *JwtManager) IsValid(token string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, errors.New("invalid JWT format")
	}
	expectedSig := j.GetSignature(parts[0], parts[1])
	if parts[2] != expectedSig && parts[2] != expectedSig+"=" {
		return false, errors.New("invalid JWT signature")
	}
	return true, nil
}

func (j *JwtManager) IsOnTime(token string) (bool, error) {
	payload, err := j.DecodePayload(token)
	if err != nil {
		return false, errors.New("Illegal base64")
	}
	exp, ok := payload["exp"].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return false, errors.New("JWT expired")
	}
	return true, nil
}

func (j *JwtManager) TokenNeedsRefresh(token string) (bool, error) {
	payload, err := j.DecodePayload(token)
	if err != nil {
		return false, errors.New("Illegal base64")
	}
	iat, ok := payload["iat"].(float64)
	if !ok {
		return false, errors.New("invalid JWT payload: missing iat")
	}
	if time.Now().Unix() > int64(iat)+j.Renew {
		return true, nil
	}
	return false, nil
}

func (j *JwtManager) DecodePayload(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	payloadJson, err := Base64UrlDecode(parts[1])
	if err != nil {
		return nil, errors.New("Illegal base64")
	}
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(payloadJson), &payload)
	if err != nil {
		return nil, errors.New("Illegal payload")
	}
	return payload, nil
}

func Base64UrlEncode(data []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}

func Base64UrlDecode(data string) (string, error) {
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(data)
	return string(decoded), err
}
