package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrInvalidJSON      = errors.New("invalid JSON document")
	ErrEmptyJSON       = errors.New("JSON document is empty")
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
	if arg == "" {
		return nil, ErrEmptyJSON
	}

	if arg == "-" {
		return ParseJSON(os.Stdin)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(arg), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return result, nil
}

func GetStringField(doc map[string]interface{}, field string) (string, bool) {
	val, exists := doc[field]
	if !exists {
		return "", false
	}

	strVal, ok := val.(string)
	if !ok {
		return "", false
	}

	trimmed := strings.TrimSpace(strVal)
	return trimmed, trimmed != ""
}

func GetStringFieldTrim(doc map[string]interface{}, field string) (string, bool) {
	val, exists := doc[field]
	if !exists {
		return "", false
	}

	strVal, ok := val.(string)
	if !ok {
		return "", false
	}

	return strVal, true
}

func GetNumberField(doc map[string]interface{}, field string) (float64, bool) {
	val, exists := doc[field]
	if !exists {
		return 0, false
	}

	numVal, ok := val.(float64)
	if !ok {
		return 0, false
	}

	return numVal, true
}

func GetBooleanField(doc map[string]interface{}, field string) (bool, bool) {
	val, exists := doc[field]
	if !exists {
		return false, false
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false, false
	}

	return boolVal, true
}