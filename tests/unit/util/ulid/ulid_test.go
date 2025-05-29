package ulid

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/not-empty/grit/app/util/ulid"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenerator_Generate_IsValidFormat(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	id, err := dg.Generate(0)
	require.NoError(t, err)
	require.Equal(t, ulid.TIME_LENGTH+ulid.RANDOM_LENGTH, len(id), "ULID length mismatch")
	require.True(t, dg.IsValidFormat(id), "Generated ULID should be recognized as valid")

	for _, r := range id {
		require.Contains(t, ulid.CHARS, string(r), "Invalid character in ULID: %c", r)
	}
}

func TestDefaultGenerator_Generate_DeterministicTimePart(t *testing.T) {
	ts := time.Date(2025, time.May, 29, 12, 0, 0, 0, time.UTC).UnixNano() / 1e6
	dg := ulid.NewDefaultGenerator()
	id1, err := dg.Generate(ts)
	require.NoError(t, err)
	id2, err := dg.Generate(ts)
	require.NoError(t, err)

	require.Equal(t, id1[:ulid.TIME_LENGTH], id2[:ulid.TIME_LENGTH])

	require.NotEqual(t, id1[ulid.TIME_LENGTH:], id2[ulid.TIME_LENGTH:])
}

func TestDefaultGenerator_IsValidFormat_Errors(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	require.False(t, dg.IsValidFormat("short"))
	long := strings.Repeat("0", ulid.TIME_LENGTH+ulid.RANDOM_LENGTH+1)
	require.False(t, dg.IsValidFormat(long))
}

func TestDecodeTime_ValidAndInvalid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	ts := time.Now().UnixNano() / 1e6
	id, err := dg.Generate(ts)
	require.NoError(t, err)
	decoded, err := dg.DecodeTime(id[:ulid.TIME_LENGTH])
	require.NoError(t, err)
	require.Equal(t, ts, decoded, "Decoded timestamp must match input timestamp")

	_, err = dg.DecodeTime("12345")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid time part length")

	bad := id[:ulid.TIME_LENGTH-1] + "!"
	_, err = dg.DecodeTime(bad)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ULID character")

	zz := strings.Repeat("Z", ulid.TIME_LENGTH)
	_, err = dg.DecodeTime(zz)
	require.Error(t, err)
	require.Contains(t, err.Error(), "timestamp too large")
}

func TestGetTimeAndDateFromUlid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	id, err := dg.Generate(0)
	require.NoError(t, err)

	tVal, err := dg.GetTimeFromUlid(id)
	require.NoError(t, err)
	now := time.Now().UnixNano() / 1e6
	require.LessOrEqual(t, tVal, now)

	dateStr, err := dg.GetDateFromUlid(id)
	require.NoError(t, err)
	require.NotEmpty(t, dateStr)
	parts := strings.Split(dateStr, " ")
	require.Len(t, parts, 2)
}

func TestGetTimeFromUlid_Invalid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	tVal, err := dg.GetTimeFromUlid("bad")
	require.Error(t, err)
	require.Zero(t, tVal)
	require.Contains(t, err.Error(), "invalid ULID format")
}

func TestGetDateFromUlid_Invalid(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	d, err := dg.GetDateFromUlid("bad")
	require.Error(t, err)
	require.Empty(t, d)
	require.Contains(t, err.Error(), "invalid ULID format")
}

func TestGetRandomnessFromString(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	id, err := dg.Generate(0)
	require.NoError(t, err)

	randPart, err := dg.GetRandomnessFromString(id)
	require.NoError(t, err)
	require.Len(t, randPart, ulid.RANDOM_LENGTH)

	_, err = dg.GetRandomnessFromString("short")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid ULID format")
}

func TestStubs_IsDuplicatedTimeAndHasIncrement(t *testing.T) {
	dg := ulid.NewDefaultGenerator()
	require.False(t, dg.IsDuplicatedTime(0))
	require.False(t, dg.HasIncrementLastRandChars(true))
	require.False(t, dg.HasIncrementLastRandChars(false))
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestGenerate_RandReaderError(t *testing.T) {
	orig := ulid.RandReader
	defer func() { ulid.RandReader = orig }()
	ulid.RandReader = &errReader{}

	dg := ulid.NewDefaultGenerator()
	id, err := dg.Generate(0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "read error")
	require.Empty(t, id, "On error, generated ULID string should be empty")
}
