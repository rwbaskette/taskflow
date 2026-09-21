package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rwbaskette/taskflow/internal/clierr"
)

var (
	ErrInvalidJSON = errors.New("invalid JSON document")
	ErrEmptyJSON   = errors.New("JSON document is empty")
)

func ParseJSON(r io.Reader) (map[string]interface{}, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return result, nil
}

func ParseJSONFromArg(arg string) (map[string]interface{}, error) {
	if arg == "-" {
		return ParseJSON(os.Stdin)
	}

	return ParseJSON(strings.NewReader(arg))
}

// getField is the shared exists + type-assert core for the typed getters.
func getField[T any](doc map[string]interface{}, field string) (T, bool) {
	var zero T
	val, exists := doc[field]
	if !exists {
		return zero, false
	}

	typed, ok := val.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// GetStringFieldTrim returns the trimmed value of a string field. It reports
// false when the field is missing, is not a string, or trims to empty.
func GetStringFieldTrim(doc map[string]interface{}, field string) (string, bool) {
	strVal, ok := getField[string](doc, field)
	if !ok {
		return "", false
	}

	trimmed := strings.TrimSpace(strVal)
	return trimmed, trimmed != ""
}

// GetNumberField returns a numeric field value.
func GetNumberField(doc map[string]interface{}, field string) (float64, bool) {
	return getField[float64](doc, field)
}

// GetIDField extracts the required 'id' parameter from a parsed JSON
// document: a missing field yields MissingIDError, a non-string value yields
// NonStringIDError, a value that trims to empty yields EmptyIDError, and
// success yields the trimmed ID.
func GetIDField(doc map[string]interface{}) (string, error) {
	if _, exists := doc["id"]; !exists {
		return "", clierr.MissingIDError()
	}

	idStr, isString := doc["id"].(string)
	if !isString {
		return "", clierr.NonStringIDError(doc["id"])
	}

	id := strings.TrimSpace(idStr)
	if id == "" {
		return "", clierr.EmptyIDError()
	}

	return id, nil
}
