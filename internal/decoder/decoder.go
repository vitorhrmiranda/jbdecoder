package decoder

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

const (
	base64BlockSize = 4
	validBase64Mod  = 0
)

// IsBase64 checks if a string is valid Base64 encoded
func IsBase64(s string) bool {
	// base64 strings should have a length that's a multiple of base64blocksize
	if len(s)%base64BlockSize != validBase64Mod {
		return false
	}

	// try to decode the string
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(s string) bool {
	var temp any
	return json.Unmarshal([]byte(s), &temp) == nil
}

// DecodeBase64String attempts decode a Base64 string and parse as JSON if valid
func DecodeBase64String(s string) any {
	if !IsBase64(s) {
		return s
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}

	// Check if the decoded data is valid UTF-8 text
	if !utf8.Valid(decoded) {
		// If it's not valid UTF-8, return the original Base64 string unchanged
		return s
	}

	decodedStr := strings.TrimSpace(string(decoded))

	// Check if the decoded string is valid JSON
	if IsValidJSON(decodedStr) {
		var jsonObj any
		if err := json.Unmarshal([]byte(decodedStr), &jsonObj); err == nil {
			// Recursively process the parsed JSON to decode any nested Base64
			return DecodeBase64Fields(jsonObj)
		}
	}

	return decodedStr
}

// DecodeBase64InMap processes all values in a map
func DecodeBase64InMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range m {
		result[key] = DecodeBase64Fields(value)
	}
	return result
}

// DecodeBase64InSlice processes all values in a slice
func DecodeBase64InSlice(s []any) []any {
	result := make([]any, len(s))
	for i, value := range s {
		result[i] = DecodeBase64Fields(value)
	}
	return result
}

// DecodeBase64Fields recursively traverses JSON data and decodes Base64 strings
func DecodeBase64Fields(data any) any {
	switch v := data.(type) {
	case map[string]any:
		return DecodeBase64InMap(v)
	case []any:
		return DecodeBase64InSlice(v)
	case string:
		return DecodeBase64String(v)
	default:
		// For other types (numbers, booleans, null), return as-is
		return v
	}
}
