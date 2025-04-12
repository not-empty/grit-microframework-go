package ulid

import (
	"errors"
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

type Ulid struct {
	LastGenTime   int64
	LastRandChars []int
}

func (u *Ulid) IsValidFormat(ulidStr string) bool {
	return len(ulidStr) == TIME_LENGTH+RANDOM_LENGTH
}

func (u *Ulid) GetTimeFromUlid(ulidStr string) (int64, error) {
	if !u.IsValidFormat(ulidStr) {
		return 0, errors.New("invalid ULID format")
	}
	timePart := ulidStr[:TIME_LENGTH]
	return u.DecodeTime(timePart)
}

func (u *Ulid) GetDateFromUlid(ulidStr string) (string, error) {
	t, err := u.GetTimeFromUlid(ulidStr)
	if err != nil {
		return "", err
	}
	return time.Unix(t, 0).Format("2006-01-02 15:04:05"), nil
}

func (u *Ulid) GetRandomnessFromString(ulidStr string) (string, error) {
	if !u.IsValidFormat(ulidStr) {
		return "", errors.New("invalid ULID format")
	}
	return ulidStr[TIME_LENGTH:], nil
}

func (u *Ulid) IsDuplicatedTime(t int64) bool {
	return t == u.LastGenTime
}

func (u *Ulid) HasIncrementLastRandChars(duplicateTime bool) bool {
	if !duplicateTime {
		u.LastRandChars = make([]int, RANDOM_LENGTH)
		for i := 0; i < RANDOM_LENGTH; i++ {
			u.LastRandChars[i] = randomInt(0, BASE-1)
		}
		return false
	}
	for i := RANDOM_LENGTH - 1; i >= 0; i-- {
		if u.LastRandChars[i] == BASE-1 {
			u.LastRandChars[i] = 0
		} else {
			u.LastRandChars[i]++
			break
		}
	}
	return true
}

func (u *Ulid) Generate(t int64) string {
	if t == 0 {
		t = int64(time.Now().UnixNano() / 1e6)
	}

	duplicateTime := u.IsDuplicatedTime(t)
	u.LastGenTime = t

	timeChars := ""
	temp := t
	for i := 0; i < TIME_LENGTH; i++ {
		mod := temp % BASE
		timeChars = string(CHARS[mod]) + timeChars
		temp = temp / BASE
	}

	u.HasIncrementLastRandChars(duplicateTime)
	randChars := ""
	for i := 0; i < RANDOM_LENGTH; i++ {
		randChars += string(CHARS[u.LastRandChars[i]])
	}

	return timeChars + randChars
}

func (u *Ulid) DecodeTime(timePart string) (int64, error) {
	if len(timePart) != TIME_LENGTH {
		return 0, errors.New("invalid time part length")
	}
	reversed := reverseString(timePart)
	var carry int64 = 0
	for i, char := range reversed {
		index := strings.IndexRune(CHARS, char)
		if index == -1 {
			return 0, errors.New("invalid ULID character: " + string(char))
		}
		carry += int64(index) * int64(math.Pow(float64(BASE), float64(i)))
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

func randomInt(min, max int) int {
	return min + int(time.Now().UnixNano()%int64(max-min+1))
}
