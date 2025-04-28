package helper

import (
	"testing"
	"time"

	"github.com/not-empty/grit/app/helper"
)

func TestUnmarshalJSON(t *testing.T) {
	var jt helper.JSONTime
	tests := []struct {
		input   string
		want    time.Time
		wantNil bool
		wantErr bool
	}{
		{`"2025-04-04T01:03:02Z"`, time.Date(2025, 4, 4, 1, 3, 2, 0, time.UTC), false, false},
		{`"2025-04-04 01:03:02"`, time.Date(2025, 4, 4, 1, 3, 2, 0, time.UTC), false, false},
		{`""`, time.Time{}, true, false},
		{`null`, time.Time{}, true, false},
		{`"invalid"`, time.Time{}, false, true},
	}
	for _, tt := range tests {
		jt = helper.JSONTime(time.Time{})
		err := jt.UnmarshalJSON([]byte(tt.input))
		if tt.wantErr {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("UnmarshalJSON(%s) unexpected error: %v", tt.input, err)
			continue
		}
		if tt.wantNil {
			if !time.Time(jt).IsZero() {
				t.Errorf("UnmarshalJSON(%s) expected zero time, got %v", tt.input, time.Time(jt))
			}
		} else {
			if got := time.Time(jt); !got.Equal(tt.want) {
				t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		jt      helper.JSONTime
		want    string
		wantErr bool
	}{
		{helper.JSONTime(time.Date(2025, 4, 4, 1, 3, 2, 0, time.UTC)), `"2025-04-04 01:03:02"`, false},
		{helper.JSONTime(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)), ``, true},
		{helper.JSONTime(time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC)), ``, true},
	}
	for _, tt := range tests {
		b, err := tt.jt.MarshalJSON()
		if tt.wantErr {
			if err == nil {
				t.Errorf("MarshalJSON(%v) expected error, got nil", time.Time(tt.jt))
			}
			continue
		}
		if err != nil {
			t.Errorf("MarshalJSON(%v) unexpected error: %v", time.Time(tt.jt), err)
			continue
		}
		if got := string(b); got != tt.want {
			t.Errorf("MarshalJSON(%v) = %s, want %s", time.Time(tt.jt), got, tt.want)
		}
	}
}

func TestValue(t *testing.T) {
	t0 := time.Date(2025, 4, 4, 1, 3, 2, 0, time.UTC)
	jt := helper.JSONTime(t0)
	v, err := jt.Value()
	if err != nil {
		t.Fatalf("Value() unexpected error: %v", err)
	}
	if vt, ok := v.(time.Time); !ok {
		t.Fatalf("Value() returned type %T, want time.Time", v)
	} else if !vt.Equal(t0) {
		t.Errorf("Value() = %v, want %v", vt, t0)
	}
}
