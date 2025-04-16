package ulid

import (
	"strings"
	"testing"
	"time"

	"github.com/not-empty/grit/app/util/ulid"
	"github.com/stretchr/testify/require"
)

const (
	TimeLength   = 10
	RandomLength = 16
)

func TestDefaultGenerator_IsValidFormat(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	id, err := dg.Generate(0)
	require.NoError(t, err)
	require.Equal(t, TimeLength+RandomLength, len(id), "ULID length should be %d", TimeLength+RandomLength)
	require.True(t, dg.IsValidFormat(id), "Generated ULID should be valid format")

	invalid := "abc"
	require.False(t, dg.IsValidFormat(invalid), "Short string should be invalid")
}

func TestDefaultGenerator_Generate(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	ulidStr, err := dg.Generate(0)
	require.NoError(t, err)
	require.Equal(t, TimeLength+RandomLength, len(ulidStr), "ULID should be %d characters long", TimeLength+RandomLength)
}

func TestDefaultGenerator_DecodeTime(t *testing.T) {
	dg := ulid.NewDefaultGenerator()

	ulidStr, err := dg.Generate(0)
	require.NoError(t, err)
	timePart := ulidStr[:TimeLength]

	decoded, err := dg.DecodeTime(timePart)
	require.NoError(t, err)
	now := time.Now().UnixNano() / 1e6
	require.LessOrEqual(t, decoded, now, "Decoded time should be less than or equal to current time")

	_, err = dg.DecodeTime("12345")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid time part length")

	invalidTimePart := timePart[:len(timePart)-1] + "!"
	_, err = dg.DecodeTime(invalidTimePart)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ULID character")
}

func TestDefaultGenerator_GetTimeFromUlid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	ulidStr, err := dg.Generate(0)
	require.NoError(t, err)

	tVal, err := dg.GetTimeFromUlid(ulidStr)
	require.NoError(t, err)
	now := time.Now().UnixNano() / 1e6
	require.LessOrEqual(t, tVal, now, "Extracted time should be ≤ current time")
}

func TestDefaultGenerator_GetDateFromUlid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	ulidStr, err := dg.Generate(0)
	require.NoError(t, err)

	dateStr, err := dg.GetDateFromUlid(ulidStr)
	require.NoError(t, err)
	require.NotEmpty(t, dateStr, "Date should not be empty")
	parts := strings.Split(dateStr, " ")
	require.Len(t, parts, 2, "Date should contain date and time parts")
}

func TestDefaultGenerator_GetRandomnessFromString(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	ulidStr, err := dg.Generate(0)
	require.NoError(t, err)

	randPart, err := dg.GetRandomnessFromString(ulidStr)
	require.NoError(t, err)
	require.Equal(t, RandomLength, len(randPart), "Randomness part should be %d characters", RandomLength)

	_, err = dg.GetRandomnessFromString("short")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ULID format")
}

func TestDefaultGenerator_IsDuplicatedTime(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	now := time.Now().UnixNano() / 1e6
	dg.LastGenTime = now
	require.True(t, dg.IsDuplicatedTime(now), "Should be duplicated time")
	require.False(t, dg.IsDuplicatedTime(now+1), "Time + 1 should not be considered duplicate")
}

func TestDefaultGenerator_HasIncrementLastRandChars(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	orig := make([]int, len(dg.LastRandChars))
	copy(orig, dg.LastRandChars)

	changed := dg.HasIncrementLastRandChars(false)
	require.False(t, changed, "Should return false for non-duplicate time")
	require.Equal(t, RandomLength, len(dg.LastRandChars))

	dg.LastRandChars = make([]int, RandomLength)
	for i := 0; i < RandomLength; i++ {
		dg.LastRandChars[i] = 0
	}
	changed = dg.HasIncrementLastRandChars(true)
	require.True(t, changed, "Should return true for duplicate time")
	require.Equal(t, 1, dg.LastRandChars[RandomLength-1])
}

func TestDefaultGenerator_DecodeTime_Error(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	_, err := dg.DecodeTime("12345")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid time part length")

	validTimePart := "0123456789"
	invalidTimePart := validTimePart[:len(validTimePart)-1] + "X"
	if strings.Contains(ulid.CHARS, "X") {
		invalidTimePart = validTimePart[:len(validTimePart)-1] + "!"
	}
	_, err = dg.DecodeTime(invalidTimePart)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ULID character")
}

func TestDefaultGenerator_HasIncrementLastRandChars_AllAtMax(t *testing.T) {
	dg := ulid.NewDefaultGenerator()

	now := time.Now().UnixNano() / 1e6
	dg.LastGenTime = now

	const maxVal = 31
	const randomLength = 16
	dg.LastRandChars = make([]int, randomLength)
	for i := 0; i < randomLength; i++ {
		dg.LastRandChars[i] = maxVal
	}

	changed := dg.HasIncrementLastRandChars(true)
	require.True(t, changed, "Expected duplicate branch to return true")

	for i, digit := range dg.LastRandChars {
		require.Equalf(t, 0, digit, "Expected digit at index %d to be reset to 0", i)
	}
}

func TestDefaultGenerator_DecodeTime_TimestampTooLarge(t *testing.T) {
	dg := ulid.NewDefaultGenerator()

	timePart := strings.Repeat("Z", 10)

	decoded, err := dg.DecodeTime(timePart)
	require.Error(t, err, "Expected an error when decoded time exceeds TIME_MAX")
	require.Equal(t, int64(0), decoded, "Decoded time should be zero when error occurs")
	require.Contains(t, err.Error(), "timestamp too large", "Error should indicate that timestamp is too large")
}

func TestDefaultGenerator_GetTimeFromUlid_InvalidFormat(t *testing.T) {
	dg := ulid.NewDefaultGenerator()

	tVal, err := dg.GetTimeFromUlid("abc")
	require.Error(t, err, "Expected error for invalid ULID format")
	require.Equal(t, int64(0), tVal, "Returned time should be zero on error")
	require.Contains(t, err.Error(), "invalid ULID format")
}

func TestDefaultGenerator_GetDateFromUlid_InvalidFormat(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	invalidULID := "invalid"
	date, err := dg.GetDateFromUlid(invalidULID)
	require.Error(t, err, "Expected an error for invalid ULID format")
	require.Equal(t, "", date, "Date should be empty when error occurs")
	require.Contains(t, err.Error(), "invalid ULID format")
}
