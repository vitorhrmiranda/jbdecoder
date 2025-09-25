package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	testTimeout   = 2 * time.Second
	testFilePerms = 0o600
)

const testJSON = `{
  "user": {
    "name": "Sm9obiBEb2U=",
    "email": "am9obi5kb2VAZXhhbXBsZS5jb20=",
    "age": 30,
    "active": true
  },
  "messages": [
    "SGVsbG8gV29ybGQ=",
    "VGhpcyBpcyBhIHRlc3QgbWVzc2FnZQ==",
    42,
    false
  ],
  "metadata": {
    "description": "QSBzYW1wbGUgSlNPTiBmaWxl",
    "version": "1.0.0",
    "tags": ["dGVzdA==", "ZGVtbw==", "not-base64"]
  }
}`

func TestMain(t *testing.T) {
	testCases := []struct {
		name   string
		cmd    func(t *testing.T) *exec.Cmd
		assert func(t *testing.T, output, stderr []byte, err error)
	}{
		{
			name: "direct JSON string argument",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", `{"message": "SGVsbG8gV29ybGQ=", "number": 42}`)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}
				expected := `{"message":"Hello World","number":42}`
				actual := strings.TrimSpace(string(output))
				if actual != expected {
					t.Errorf("Expected: %s, Got: %s", expected, actual)
				}
			},
		},
		{
			name: "base64 that is another JSON",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", `{"message": "eyJrZXkiOiJ2YWx1ZSJ9Cg==", "number": 42}`)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}
				expected := `{"message":{"key":"value"},"number":42}`
				actual := strings.TrimSpace(string(output))
				if actual != expected {
					t.Errorf("Expected: %s, Got: %s", expected, actual)
				}
			},
		},
		{
			name: "file input argument",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				// Create a temporary test file
				const tmpFile = "test_input.json"
				err := os.WriteFile(tmpFile, []byte(testJSON), testFilePerms)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				t.Cleanup(func() { os.Remove(tmpFile) })

				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", tmpFile)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				// Parse the output to verify it's valid JSON and has expected decodings
				var result map[string]any
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
					return
				}

				// Verify specific decodings
				user := result["user"].(map[string]any)
				if user["name"] != "John Doe" {
					t.Errorf("Expected user name 'John Doe', got '%v'", user["name"])
				}
				if user["email"] != "john.doe@example.com" {
					t.Errorf("Expected user email 'john.doe@example.com', got '%v'", user["email"])
				}

				messages := result["messages"].([]any)
				if messages[0] != "Hello World" {
					t.Errorf("Expected first message 'Hello World', got '%v'", messages[0])
				}
				if messages[1] != "This is a test message" {
					t.Errorf("Expected second message 'This is a test message', got '%v'", messages[1])
				}
			},
		},
		{
			name: "stdin redirect",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				cmd := exec.CommandContext(ctx, "go", "run", "main.go")
				cmd.Stdin = strings.NewReader(`{"data": "SGVsbG8="}`)
				return cmd
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}
				expected := `{"data":"Hello"}`
				actual := strings.TrimSpace(string(output))
				if actual != expected {
					t.Errorf("Expected: %s, Got: %s", expected, actual)
				}
			},
		},
		{
			name: "empty JSON object",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", "{}")
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}
				expected := `{}`
				actual := strings.TrimSpace(string(output))
				if actual != expected {
					t.Errorf("Expected: %s, Got: %s", expected, actual)
				}
			},
		},
		{
			name: "pipe input",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				cmd := exec.CommandContext(ctx, "go", "run", "main.go")
				cmd.Stdin = strings.NewReader(`{"encoded": "SGVsbG8=", "number": 123}`)
				return cmd
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				// Verify the output
				var result map[string]any
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
					return
				}

				if result["encoded"] != "Hello" {
					t.Errorf("Expected encoded field to be 'Hello', got '%v'", result["encoded"])
				}
				if result["number"].(float64) != 123 {
					t.Errorf("Expected number field to be 123, got '%v'", result["number"])
				}
			},
		},
		{
			name: "invalid JSON",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", `{"invalid": json}`)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err == nil {
					t.Errorf("Expected command to fail with invalid JSON")
					return
				}

				stderrOutput := string(stderr)
				if !strings.Contains(stderrOutput, "Error parsing JSON") {
					t.Errorf("Expected error message about parsing JSON, got: %s", stderrOutput)
				}
			},
		},
		{
			name: "non-existent file",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", "nonexistent.json")
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err == nil {
					t.Errorf("Expected command to fail with non-existent file")
					return
				}

				stderrOutput := string(stderr)
				if !strings.Contains(stderrOutput, "Error reading input") {
					t.Errorf("Expected error message about reading input, got: %s", stderrOutput)
				}
			},
		},
		{
			name: "complex nested JSON",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				complexJSON := `{
					"level1": {
						"encoded": "SGVsbG8=",
						"plain": "world",
						"level2": {
							"array": ["VGVzdA==", 123, true, "not-base64"],
							"nested_encoded": "V29ybGQ="
						}
					},
					"root_array": ["Rm9v", "YmFy", false, null]
				}`
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", complexJSON)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				var result map[string]any
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
					return
				}

				// Verify nested decoding
				level1 := result["level1"].(map[string]any)
				if level1["encoded"] != "Hello" {
					t.Errorf("Expected level1.encoded to be 'Hello', got '%v'", level1["encoded"])
				}

				level2 := level1["level2"].(map[string]any)
				if level2["nested_encoded"] != "World" {
					t.Errorf("Expected level2.nested_encoded to be 'World', got '%v'", level2["nested_encoded"])
				}

				array := level2["array"].([]any)
				if array[0] != "Test" {
					t.Errorf("Expected array[0] to be 'Test', got '%v'", array[0])
				}

				rootArray := result["root_array"].([]any)
				if rootArray[0] != "Foo" || rootArray[1] != "bar" {
					t.Errorf("Expected root_array to contain decoded 'Foo' and 'bar', got %v", rootArray)
				}
			},
		},
		{
			name: "base64 edge cases",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				edgeCasesJSON := `{
					"valid_base64": "SGVsbG8gV29ybGQ=",
					"not_base64_wrong_length": "abc",
					"not_base64_invalid_chars": "Hello@World",
					"empty_string": "",
					"number": 42,
					"boolean": true,
					"null_value": null,
					"array_with_mixed": ["SGVsbG8=", "plain text", 123, "VGVzdA=="]
				}`
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", edgeCasesJSON)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				var result map[string]any
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
					return
				}

				// Verify that valid Base64 is decoded
				if result["valid_base64"] != "Hello World" {
					t.Errorf("Expected valid_base64 to be 'Hello World', got '%v'", result["valid_base64"])
				}

				// Verify that invalid Base64 strings are left unchanged
				if result["not_base64_wrong_length"] != "abc" {
					t.Errorf("Expected not_base64_wrong_length to remain 'abc', got '%v'", result["not_base64_wrong_length"])
				}

				if result["not_base64_invalid_chars"] != "Hello@World" {
					t.Errorf("Expected not_base64_invalid_chars to remain 'Hello@World', got '%v'", result["not_base64_invalid_chars"])
				}

				// Verify that other types are preserved
				if result["number"].(float64) != 42 {
					t.Errorf("Expected number to be 42, got '%v'", result["number"])
				}

				if result["boolean"] != true {
					t.Errorf("Expected boolean to be true, got '%v'", result["boolean"])
				}

				// Verify mixed array
				array := result["array_with_mixed"].([]any)
				if array[0] != "Hello" { // SGVsbG8= decoded
					t.Errorf("Expected array[0] to be 'Hello', got '%v'", array[0])
				}
				if array[1] != "plain text" { // should remain unchanged
					t.Errorf("Expected array[1] to be 'plain text', got '%v'", array[1])
				}
				if array[3] != "Test" { // VGVzdA== decoded
					t.Errorf("Expected array[3] to be 'Test', got '%v'", array[3])
				}
			},
		},
		{
			name: "show help on no input",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				cmd := exec.CommandContext(ctx, "go", "run", "main.go")
				cmd.Stdin = strings.NewReader("")
				return cmd
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				outputStr := string(output)
				if !strings.Contains(outputStr, "JSON Base64 Decoder") {
					t.Errorf("Expected help message to be displayed when no input is provided")
				}
				if !strings.Contains(outputStr, "USAGE:") {
					t.Errorf("Expected help message to contain USAGE section")
				}
				if !strings.Contains(outputStr, "jbdecoder [INPUT]") {
					t.Errorf("Expected help message to contain usage pattern")
				}
			},
		},
		{
			name: "show help on empty input",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				cmd := exec.CommandContext(ctx, "go", "run", "main.go")
				cmd.Stdin = strings.NewReader("   \n  \t  ")
				return cmd
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				outputStr := string(output)
				if !strings.Contains(outputStr, "JSON Base64 Decoder") {
					t.Errorf("Expected help message to be displayed when empty input is provided")
				}
				if !strings.Contains(outputStr, "USAGE:") {
					t.Errorf("Expected help message to contain USAGE section")
				}
			},
		},
		{
			name: "base64 binary data",
			cmd: func(t *testing.T) *exec.Cmd {
				t.Helper()
				// Create a Base64 string that contains binary data (not valid UTF-8)
				binaryData := []byte{0xFF, 0xFE, 0x00, 0x01, 0x80, 0x81} // Invalid UTF-8 bytes
				binaryBase64 := base64.StdEncoding.EncodeToString(binaryData)
				testJSON := fmt.Sprintf(`{"binaryField": "%s", "textField": "SGVsbG8="}`, binaryBase64)

				ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
				t.Cleanup(cancel)
				return exec.CommandContext(ctx, "go", "run", "main.go", testJSON)
			},
			assert: func(t *testing.T, output []byte, stderr []byte, err error) {
				t.Helper()
				if err != nil {
					t.Errorf("Command failed: %v", err)
					return
				}

				var result map[string]any
				if err := json.Unmarshal(output, &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
					return
				}

				// Create the same binary data for comparison
				binaryData := []byte{0xFF, 0xFE, 0x00, 0x01, 0x80, 0x81}
				binaryBase64 := base64.StdEncoding.EncodeToString(binaryData)

				// The binary Base64 should remain unchanged (not decoded)
				if result["binaryField"] != binaryBase64 {
					t.Errorf("Expected binaryField to remain as Base64 '%s', got '%v'", binaryBase64, result["binaryField"])
				}

				// The text Base64 should be decoded normally
				if result["textField"] != "Hello" {
					t.Errorf("Expected textField to be decoded to 'Hello', got '%v'", result["textField"])
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cmd := testCase.cmd(t)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			output, err := cmd.Output()
			testCase.assert(t, output, stderr.Bytes(), err)
		})
	}
}
