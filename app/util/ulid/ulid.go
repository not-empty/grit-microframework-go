package ulid

import (
	"crypto/rand"
	"errors"
	"io"
	"math"
	"strings"
	"time"
)

const (
	CHARS         = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	BASE          = 32
	TIME_MAX      = 281474976710655
	TIME_LENGTH   = 10
	RANDOM_LENGTH = 16
)

var RandReader io.Reader = rand.Reader

type Generator interface {
	IsValidFormat(ulidStr string) bool
	GetTimeFromUlid(ulidStr string) (int64, error)
	GetDateFromUlid(ulidStr string) (string, error)
	GetRandomnessFromString(ulidStr string) (string, error)
	IsDuplicatedTime(t int64) bool
	HasIncrementLastRandChars(duplicateTime bool) bool
	Generate(ts int64) (string, error)
	DecodeTime(timePart string) (int64, error)
}

type DefaultGenerator struct{}

func NewDefaultGenerator() *DefaultGenerator {
	return &DefaultGenerator{}
}

func (u *DefaultGenerator) IsValidFormat(ulidStr string) bool {
	return len(ulidStr) == TIME_LENGTH+RANDOM_LENGTH
}

func (u *DefaultGenerator) GetTimeFromUlid(ulidStr string) (int64, error) {
	if !u.IsValidFormat(ulidStr) {
		return 0, errors.New("invalid ULID format")
	}
	return u.DecodeTime(ulidStr[:TIME_LENGTH])
}

func (u *DefaultGenerator) GetDateFromUlid(ulidStr string) (string, error) {
	t, err := u.GetTimeFromUlid(ulidStr)
	if err != nil {
		return "", err
	}
	return time.Unix(t/1000, (t%1000)*1e6).Format("2006-01-02 15:04:05"), nil
}

func (u *DefaultGenerator) GetRandomnessFromString(ulidStr string) (string, error) {
	if !u.IsValidFormat(ulidStr) {
		return "", errors.New("invalid ULID format")
	}
	return ulidStr[TIME_LENGTH:], nil
}

func (u *DefaultGenerator) IsDuplicatedTime(t int64) bool {
	return false
}

func (u *DefaultGenerator) HasIncrementLastRandChars(duplicateTime bool) bool {
	return false
}

func (u *DefaultGenerator) Generate(ts int64) (string, error) {
	if ts == 0 {
		ts = time.Now().UnixNano() / 1e6
	}

	var out [TIME_LENGTH + RANDOM_LENGTH]byte

	tmp := ts
	for i := TIME_LENGTH - 1; i >= 0; i-- {
		out[i] = CHARS[tmp%BASE]
		tmp /= BASE
	}

	var buf [RANDOM_LENGTH]byte
	if _, err := RandReader.Read(buf[:]); err != nil {
		return "", err
	}

	for i := 0; i < RANDOM_LENGTH; i++ {
		out[TIME_LENGTH+i] = CHARS[buf[i]&0x1F]
	}

	return string(out[:]), nil
}

func (u *DefaultGenerator) DecodeTime(timePart string) (int64, error) {
	if len(timePart) != TIME_LENGTH {
		return 0, errors.New("invalid time part length")
	}
	rev := reverseString(timePart)
	var carry int64
	for i, r := range rev {
		idx := strings.IndexRune(CHARS, r)
		if idx < 0 {
			return 0, errors.New("invalid ULID character: " + string(r))
		}
		carry += int64(idx) * int64(math.Pow(BASE, float64(i)))
	}
	if carry > TIME_MAX {
		return 0, errors.New("timestamp too large")
	}
	return carry, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
