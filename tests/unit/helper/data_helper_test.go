package helper

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/not-empty/grit-microframework-go/app/helper"
)

type sample struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestFilterJSON_AllFields(t *testing.T) {
	input := sample{"123", "Leo", "leo@example.com", 30}
	result := helper.FilterJSON(input, nil)

	expected := map[string]interface{}{
		"id":    "123",
		"name":  "Leo",
		"email": "leo@example.com",
		"age":   float64(30),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFilterJSON_SpecificFields(t *testing.T) {
	input := sample{"123", "Leo", "leo@example.com", 30}
	result := helper.FilterJSON(input, []string{"id", "name"})

	expected := map[string]interface{}{
		"id":   "123",
		"name": "Leo",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFilterJSON_UnknownField(t *testing.T) {
	input := sample{"123", "Leo", "leo@example.com", 30}
	result := helper.FilterJSON(input, []string{"nonexistent", "id"})

	expected := map[string]interface{}{
		"id": "123",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFilterJSON_EmptyFieldsSlice(t *testing.T) {
	input := sample{"123", "Leo", "leo@example.com", 30}
	result := helper.FilterJSON(input, []string{})

	expected := map[string]interface{}{
		"id":    "123",
		"name":  "Leo",
		"email": "leo@example.com",
		"age":   float64(30),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestIsEmptyValue_Types(t *testing.T) {
	now := time.Now()
	jsonTime := helper.JSONTime(now)

	tt := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"nil interface", nil, true},

		{"empty string", "", true},
		{"non-empty string", "foo", false},

		{"int zero", 0, true},
		{"int non-zero", 42, false},

		{"int64 zero", int64(0), true},
		{"int64 non-zero", int64(-7), false},

		{"float64 zero", float64(0), true},
		{"float64 non-zero", float64(3.14), false},

		{"*time.Time nil", (*time.Time)(nil), true},
		{"*time.Time non-nil", &now, false},

		{"*JSONTime nil", (*helper.JSONTime)(nil), true},
		{"*JSONTime non-nil", &jsonTime, false},

		{"bool (default case)", true, false},

		{"slice (default case)", []int{}, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.IsEmptyValue(tc.input)
			if got != tc.want {
				t.Errorf("IsEmptyValue(%#v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestFilterOutDefaulted_NoDefaults(t *testing.T) {
	cols := []string{"id", "name", "age"}
	vals := []interface{}{"123", "Leo", 0}
	fCols, fVals := helper.FilterOutDefaulted(cols, vals, nil)

	if !reflect.DeepEqual(fCols, cols) {
		t.Errorf("FilterOutDefaulted no defaults, got cols %v, want %v", fCols, cols)
	}
	if !reflect.DeepEqual(fVals, vals) {
		t.Errorf("FilterOutDefaulted no defaults, got vals %v, want %v", fVals, vals)
	}
}

func TestFilterOutDefaulted_WithDefaultsAndEmptyValues(t *testing.T) {
	cols := []string{"id", "name", "age", "email"}
	vals := []interface{}{"123", "", 0, "leo@example.com"}
	defaultCols := []string{"name", "age"}

	wantCols := []string{"id", "email"}
	wantVals := []interface{}{"123", "leo@example.com"}

	fCols, fVals := helper.FilterOutDefaulted(cols, vals, defaultCols)
	if !reflect.DeepEqual(fCols, wantCols) {
		t.Errorf("FilterOutDefaulted(cols=%v, vals=%v, defaultCols=%v) => cols %v, want %v",
			cols, vals, defaultCols, fCols, wantCols)
	}
	if !reflect.DeepEqual(fVals, wantVals) {
		t.Errorf("FilterOutDefaulted(vals) => %v, want %v", fVals, wantVals)
	}
}

func TestFilterOutDefaulted_DefaultColsButNonEmpty(t *testing.T) {
	cols := []string{"id", "name", "age"}
	vals := []interface{}{"123", "Alice", 0}
	defaultCols := []string{"name", "age"}

	wantCols := []string{"id", "name"}
	wantVals := []interface{}{"123", "Alice"}

	fCols, fVals := helper.FilterOutDefaulted(cols, vals, defaultCols)
	if !reflect.DeepEqual(fCols, wantCols) {
		t.Errorf("FilterOutDefaulted => cols %v, want %v", fCols, wantCols)
	}
	if !reflect.DeepEqual(fVals, wantVals) {
		t.Errorf("FilterOutDefaulted => vals %v, want %v", fVals, wantVals)
	}
}

func TestFilterOutDefaulted_AllFilteredOut(t *testing.T) {
	cols := []string{"a", "b"}
	vals := []interface{}{"", 0}
	defaultCols := []string{"a", "b"}

	fCols, fVals := helper.FilterOutDefaulted(cols, vals, defaultCols)
	if len(fCols) != 0 || len(fVals) != 0 {
		t.Errorf("FilterOutDefaulted => cols %v, vals %v, want both empty slices", fCols, fVals)
	}
}

func TestFilterOutDefaulted_NoMatchesInDefaultCols(t *testing.T) {
	cols := []string{"id", "value"}
	vals := []interface{}{"123", 3.14}
	defaultCols := []string{"foo", "bar"}

	wantCols := []string{"id", "value"}
	wantVals := []interface{}{"123", 3.14}

	fCols, fVals := helper.FilterOutDefaulted(cols, vals, defaultCols)
	if !reflect.DeepEqual(fCols, wantCols) {
		t.Errorf("FilterOutDefaulted => cols %v, want %v", fCols, wantCols)
	}
	if !reflect.DeepEqual(fVals, wantVals) {
		t.Errorf("FilterOutDefaulted => vals %v, want %v", fVals, wantVals)
	}
}

func TestBuildRowTokens_NoDefaults(t *testing.T) {
	allCols := []string{"a", "b"}
	vals := []interface{}{"foo", "bar"}
	defaultCols := []string{}

	rowSQL, args := helper.BuildRowTokens(allCols, vals, defaultCols)

	wantSQL := "(?, ?)"
	wantArgs := []interface{}{"foo", "bar"}

	if rowSQL != wantSQL {
		t.Errorf("BuildRowTokens no defaults: got SQL %q, want %q", rowSQL, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("BuildRowTokens no defaults: got args %v, want %v", args, wantArgs)
	}
}

func TestBuildRowTokens_DefaultEmptyInt(t *testing.T) {
	allCols := []string{"a", "b", "c"}
	vals := []interface{}{"foo", 0, "baz"}
	defaultCols := []string{"b"}

	rowSQL, args := helper.BuildRowTokens(allCols, vals, defaultCols)

	wantSQL := "(?, DEFAULT, ?)"
	wantArgs := []interface{}{"foo", "baz"}

	if rowSQL != wantSQL {
		t.Errorf("BuildRowTokens default empty int: got SQL %q, want %q", rowSQL, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("BuildRowTokens default empty int: got args %v, want %v", args, wantArgs)
	}
}

func TestBuildRowTokens_DefaultEmptyString(t *testing.T) {
	allCols := []string{"x", "y"}
	vals := []interface{}{"", "hello"}
	defaultCols := []string{"x"}

	rowSQL, args := helper.BuildRowTokens(allCols, vals, defaultCols)

	wantSQL := "(DEFAULT, ?)"
	wantArgs := []interface{}{"hello"}

	if rowSQL != wantSQL {
		t.Errorf("BuildRowTokens default empty string: got SQL %q, want %q", rowSQL, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("BuildRowTokens default empty string: got args %v, want %v", args, wantArgs)
	}
}

func TestBuildRowTokens_DefaultEmptyPointerAndFloat(t *testing.T) {
	var pt *time.Time = nil
	allCols := []string{"a", "b", "c"}
	vals := []interface{}{"v", pt, float64(0)}
	defaultCols := []string{"b", "c"}

	rowSQL, args := helper.BuildRowTokens(allCols, vals, defaultCols)

	wantSQL := "(?, DEFAULT, DEFAULT)"
	wantArgs := []interface{}{"v"}

	if rowSQL != wantSQL {
		t.Errorf("BuildRowTokens default empty pointer/float: got SQL %q, want %q", rowSQL, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("BuildRowTokens default empty pointer/float: got args %v, want %v", args, wantArgs)
	}
}

func TestBuildRowTokens_DefaultNonEmptyPointer(t *testing.T) {
	now := time.Now()
	pt := &now

	allCols := []string{"a", "b"}
	vals := []interface{}{"", pt}
	defaultCols := []string{"a", "b"}

	rowSQL, args := helper.BuildRowTokens(allCols, vals, defaultCols)

	wantSQL := "(DEFAULT, ?)"
	wantArgs := []interface{}{pt}

	if rowSQL != wantSQL {
		t.Errorf("BuildRowTokens default non-empty pointer: got SQL %q, want %q", rowSQL, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("BuildRowTokens default non-empty pointer: got args %v, want %v", args, wantArgs)
	}
}

func TestBuildRowTokens_OrderAndSpacing(t *testing.T) {
	allCols := []string{"col1", "col2", "col3"}
	vals := []interface{}{1, "", 3}
	defaultCols := []string{"col2"}

	rowSQL, _ := helper.BuildRowTokens(allCols, vals, defaultCols)

	if !strings.HasPrefix(rowSQL, "(") || !strings.HasSuffix(rowSQL, ")") {
		t.Errorf("BuildRowTokens order/spacing: rowSQL %q missing surrounding parens", rowSQL)
	}
	parts := strings.Split(strings.Trim(rowSQL, "()"), ", ")
	if len(parts) != 3 {
		t.Errorf("BuildRowTokens order/spacing: expected 3 tokens, got %v", parts)
	}
	if parts[0] != "?" || parts[1] != "DEFAULT" || parts[2] != "?" {
		t.Errorf("BuildRowTokens order/spacing: got tokens %v, want [\"?\",\"DEFAULT\",\"?\"]", parts)
	}
}
