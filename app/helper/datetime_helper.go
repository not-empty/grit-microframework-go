package helper

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type JSONTime time.Time

func (jt *JSONTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	var parseErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			*jt = JSONTime(t)
			return nil
		} else {
			parseErr = err
		}
	}
	return fmt.Errorf("JSONTime: Invalid format %q: %w", s, parseErr)
}

func (jt JSONTime) MarshalJSON() ([]byte, error) {
	t := time.Time(jt)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, fmt.Errorf("JSONTime: Year out of reach: %d", y)
	}
	return []byte(`"` + t.Format("2006-01-02 15:04:05") + `"`), nil
}

func (jt JSONTime) Value() (driver.Value, error) {
	return time.Time(jt), nil
}
