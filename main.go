package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	base64BlockSize = 4
	noArgs          = 0
	oneArg          = 1
	exitCodeError   = 1
	firstElement    = 0
	validBase64Mod  = 0
)

// isBase64 checks if a string is valid Base64 encoded
func isBase64(s string) bool {
	// Base64 strings should have a length that's a multiple of base64BlockSize
	if len(s)%base64BlockSize != validBase64Mod {
		return false
	}

	// Try to decode the string
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// isValidUTF8 checks if the byte slice contains valid UTF-8 encoded text
func isValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var temp any
	return json.Unmarshal([]byte(s), &temp) == nil
}

// decodeBase64String attempts to decode a Base64 string and parse as JSON if valid
func decodeBase64String(s string) any {
	if !isBase64(s) {
		return s
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}

	// Check if the decoded data is valid UTF-8 text
	if !isValidUTF8(decoded) {
		// If it's not valid UTF-8, return the original Base64 string unchanged
		return s
	}

	decodedStr := strings.TrimSpace(string(decoded))

	// Check if the decoded string is valid JSON
	if isValidJSON(decodedStr) {
		var jsonObj any
		if err := json.Unmarshal([]byte(decodedStr), &jsonObj); err == nil {
			// Recursively process the parsed JSON to decode any nested Base64
			return decodeBase64Fields(jsonObj)
		}
	}

	return decodedStr
}

// decodeBase64InMap processes all values in a map
func decodeBase64InMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range m {
		result[key] = decodeBase64Fields(value)
	}
	return result
}

// decodeBase64InSlice processes all values in a slice
func decodeBase64InSlice(s []any) []any {
	result := make([]any, len(s))
	for i, value := range s {
		result[i] = decodeBase64Fields(value)
	}
	return result
}

// decodeBase64Fields recursively traverses JSON data and decodes Base64 strings
func decodeBase64Fields(data any) any {
	switch v := data.(type) {
	case map[string]any:
		return decodeBase64InMap(v)
	case []any:
		return decodeBase64InSlice(v)
	case string:
		return decodeBase64String(v)
	default:
		// For other types (numbers, booleans, null), return as-is
		return v
	}
}

// showUsage displays the help message
func showUsage() {
	programName := "jbdecoder"

	_, _ = fmt.Printf(`JSON Base64 Decoder

A command-line utility to recursively decode Base64 encoded strings in JSON.

USAGE:
    %s [OPTIONS] [INPUT]

INPUT METHODS:
    # Read from stdin (pipe)
    echo '{"data": "SGVsbG8="}' | %s

    # Read from file (redirect)
    %s < input.json

    # Read from file (argument)
    %s input.json

    # Direct JSON string argument
    %s '{"message": "SGVsbG8gV29ybGQ="}'

DESCRIPTION:
    This tool recursively traverses JSON data and decodes any string fields
    that contain valid Base64 encoded data. Other data types (numbers,
    booleans, arrays, objects) are preserved unchanged.

    Only strings that are valid Base64 will be decoded. Invalid Base64
    strings are left unchanged.

OPTIONS:
    -h, --help    Show this help message and exit

EXAMPLES:
    # Decode Base64 strings in a JSON file
    %s data.json

    # Decode from stdin
    cat data.json | %s

    # Decode a simple JSON string
    %s '{"name": "Sm9obg==", "age": 30}'

    # Handle complex nested JSON
    %s '{"user": {"token": "dG9rZW4="}, "items": ["aXRlbTE="]}'

ERROR HANDLING:
    The program will exit with code 1 and display an error message to stderr:
    - Input JSON is malformed
    - File cannot be read
    - Too many arguments are provided

OUTPUT:
    The decoded JSON is written to stdout in compact format.

`, programName, programName, programName, programName, programName,
		programName, programName, programName, programName)
} // isStdinEmpty checks if stdin has no data available
func isStdinEmpty() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

// getJSONInput reads JSON input from various sources
func getJSONInput() ([]byte, error) {
	args := flag.Args()

	switch len(args) {
	case noArgs:
		// No arguments - check if stdin has data
		if isStdinEmpty() {
			return nil, errors.New("no input provided")
		}
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		// Check if stdin is empty after reading
		if len(strings.TrimSpace(string(data))) == 0 {
			return nil, errors.New("empty input provided")
		}
		return data, nil

	case oneArg:
		arg := args[firstElement]

		// Check if the argument looks like a JSON string (starts with { or [)
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, "{") || strings.HasPrefix(arg, "[") {
			// Direct JSON string argument
			return []byte(arg), nil
		}

		// Otherwise, treat it as a filename
		file, err := os.Open(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to open file '%s': %w", arg, err)
		}
		defer file.Close()

		return io.ReadAll(file)

	default:
		return nil, errors.New("too many arguments provided")
	}
}

func main() {
	// Setup flags
	help := flag.Bool("h", false, "Show help message")
	flag.BoolVar(help, "help", false, "Show help message")

	// Custom usage function
	flag.Usage = showUsage

	// Parse flags
	flag.Parse()

	// Show help if requested
	if *help {
		showUsage()
		return
	}

	// Read JSON input
	jsonData, err := getJSONInput()
	if err != nil {
		// Check if error is due to no input or empty input - show help instead of error
		if strings.Contains(err.Error(), "no input provided") || strings.Contains(err.Error(), "empty input provided") {
			showUsage()
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(exitCodeError)
	}

	// Parse JSON
	var data any
	if parseErr := json.Unmarshal(jsonData, &data); parseErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", parseErr)
		os.Exit(exitCodeError)
	}

	// Process the JSON data to decode Base64 fields
	processedData := decodeBase64Fields(data)

	// Convert back to JSON and output
	output, err := json.Marshal(processedData)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error generating output JSON: %v\n", err)
		os.Exit(exitCodeError)
	}

	_, _ = fmt.Println(string(output))
}
