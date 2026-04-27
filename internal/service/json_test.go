package service

import (
	"strings"
	"testing"
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

func TestGetStringField_Exists(t *testing.T) {
	doc := map[string]interface{}{"id": "  1  ", "title": "Test"}
	id, ok := GetStringField(doc, "id")
	if !ok {
		t.Error("expected ok to be true")
	}
	if id != "1" {
		t.Errorf("expected trimmed id '1', got %q", id)
	}
}

func TestGetStringField_NotExists(t *testing.T) {
	doc := map[string]interface{}{"id": "1"}
	_, ok := GetStringField(doc, "title")
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestGetStringField_WrongType(t *testing.T) {
	doc := map[string]interface{}{"id": 123}
	_, ok := GetStringField(doc, "id")
	if ok {
		t.Error("expected ok to be false for non-string type")
	}
}

func TestGetStringField_EmptyString(t *testing.T) {
	doc := map[string]interface{}{"id": "   "}
	val, ok := GetStringField(doc, "id")
	if ok {
		t.Error("expected ok to be false for whitespace-only string")
	}
	if val != "" {
		t.Errorf("expected empty value, got %q", val)
	}
}

func TestGetStringFieldTrim_Exists(t *testing.T) {
	doc := map[string]interface{}{"id": "  1  "}
	id, ok := GetStringFieldTrim(doc, "id")
	if !ok {
		t.Error("expected ok to be true")
	}
	if id != "  1  " {
		t.Errorf("expected original id '  1  ', got %q", id)
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
	doc := map[string]interface{}{"id": ""}
	id, ok := GetStringFieldTrim(doc, "id")
	if !ok {
		t.Error("expected ok to be true even for empty string")
	}
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}