package service

import (
	"errors"
	"strings"
	"testing"

	cliErrors "github.com/rwbaskette/taskflow/internal/errors"
)

func TestParseJSONFromArg_ValidJSON(t *testing.T) {
	input := `{"id":"1","title":"Test","milestone":"v1"}`
	doc, err := ParseJSONFromArg(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc["id"] != "1" {
		t.Errorf("expected id '1', got %v", doc["id"])
	}
	if doc["title"] != "Test" {
		t.Errorf("expected title 'Test', got %v", doc["title"])
	}
	if doc["milestone"] != "v1" {
		t.Errorf("expected milestone 'v1', got %v", doc["milestone"])
	}
}

func TestParseJSONFromArg_EmptyString(t *testing.T) {
	_, err := ParseJSONFromArg("")
	if err != ErrEmptyJSON {
		t.Errorf("expected ErrEmptyJSON, got %v", err)
	}
}

func TestParseJSONFromArg_InvalidJSON(t *testing.T) {
	_, err := ParseJSONFromArg("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected invalid JSON error, got %v", err)
	}
}

func TestGetStringFieldTrim_Exists(t *testing.T) {
	doc := map[string]interface{}{"id": "  1  "}
	id, ok := GetStringFieldTrim(doc, "id")
	if !ok {
		t.Error("expected ok to be true")
	}
	if id != "1" {
		t.Errorf("expected trimmed id '1', got %q", id)
	}
}

func TestGetStringFieldTrim_NotExists(t *testing.T) {
	doc := map[string]interface{}{"id": "1"}
	_, ok := GetStringFieldTrim(doc, "title")
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetStringFieldTrim_WrongType(t *testing.T) {
	doc := map[string]interface{}{"id": 123}
	_, ok := GetStringFieldTrim(doc, "id")
	if ok {
		t.Error("expected ok to be false for non-string type")
	}
}

func TestGetStringFieldTrim_EmptyString(t *testing.T) {
	doc := map[string]interface{}{"id": "   "}
	val, ok := GetStringFieldTrim(doc, "id")
	if ok {
		t.Error("expected ok to be false for whitespace-only string")
	}
	if val != "" {
		t.Errorf("expected empty value, got %q", val)
	}
}

func TestGetStringFieldTrim_PlainValue(t *testing.T) {
	doc := map[string]interface{}{"id": ""}
	id, ok := GetStringFieldTrim(doc, "id")
	if ok {
		t.Error("expected ok to be false for empty string")
	}
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestGetNumberField(t *testing.T) {
	doc := map[string]interface{}{"limit": float64(5), "bad": "not-a-number"}

	v, ok := GetNumberField(doc, "limit")
	if !ok || v != 5 {
		t.Errorf("expected (5, true), got (%v, %v)", v, ok)
	}

	if _, ok := GetNumberField(doc, "bad"); ok {
		t.Error("expected ok to be false for non-number type")
	}

	if _, ok := GetNumberField(doc, "missing"); ok {
		t.Error("expected ok to be false for missing field")
	}
}

func TestGetIDField_Missing(t *testing.T) {
	_, err := GetIDField(map[string]interface{}{"title": "Test"})
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}

	want := cliErrors.MissingIDError()
	var got *cliErrors.CLIError
	if !errors.As(err, &got) {
		t.Fatalf("expected CLIError, got %v", err)
	}
	if got.Code != want.Code {
		t.Errorf("expected code %v, got %v", want.Code, got.Code)
	}
	if got.Message != want.Message {
		t.Errorf("expected message %q, got %q", want.Message, got.Message)
	}
}

func TestGetIDField_NonString(t *testing.T) {
	_, err := GetIDField(map[string]interface{}{"id": 123})
	if err == nil {
		t.Fatal("expected error for non-string id, got nil")
	}

	want := cliErrors.NonStringIDError(123)
	var got *cliErrors.CLIError
	if !errors.As(err, &got) {
		t.Fatalf("expected CLIError, got %v", err)
	}
	if got.Code != want.Code {
		t.Errorf("expected code %v, got %v", want.Code, got.Code)
	}
	if got.Message != want.Message {
		t.Errorf("expected message %q, got %q", want.Message, got.Message)
	}
	if got.InvalidParams["id"] != want.InvalidParams["id"] {
		t.Errorf("expected invalid params %v, got %v", want.InvalidParams, got.InvalidParams)
	}
}

func TestGetIDField_Empty(t *testing.T) {
	_, err := GetIDField(map[string]interface{}{"id": "   "})
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}

	want := cliErrors.EmptyIDError()
	var got *cliErrors.CLIError
	if !errors.As(err, &got) {
		t.Fatalf("expected CLIError, got %v", err)
	}
	if got.Code != want.Code {
		t.Errorf("expected code %v, got %v", want.Code, got.Code)
	}
	if got.Message != want.Message {
		t.Errorf("expected message %q, got %q", want.Message, got.Message)
	}
}

func TestGetIDField_OK(t *testing.T) {
	id, err := GetIDField(map[string]interface{}{"id": "  task-42  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "task-42" {
		t.Errorf("expected trimmed id 'task-42', got %q", id)
	}
}
