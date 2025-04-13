package helper

import (
	"reflect"
	"testing"

	"github.com/not-empty/grit/app/helper"
)

type Sample struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestFilterJSON_AllFields(t *testing.T) {
	input := Sample{"123", "Leo", "leo@example.com", 30}
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
	input := Sample{"123", "Leo", "leo@example.com", 30}
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
	input := Sample{"123", "Leo", "leo@example.com", 30}
	result := helper.FilterJSON(input, []string{"nonexistent", "id"})

	expected := map[string]interface{}{
		"id": "123",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFilterJSON_EmptyFields(t *testing.T) {
	input := Sample{"123", "Leo", "leo@example.com", 30}
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
