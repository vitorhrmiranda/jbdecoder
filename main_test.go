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
	testTimeout   = 30 * time.Second
	testFilePerms = 0o600
)

// TestDirectJSONStringArgument tests passing JSON directly as command line argument
func TestDirectJSONStringArgument(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", `{"message": "SGVsbG8gV29ybGQ=", "number": 42}`)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	expected := `{"message":"Hello World","number":42}`
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Errorf("Expected: %s, Got: %s", expected, actual)
	}
}

// TestBase64IsAnotherJSON tests decoding Base64 that is itself a JSON string
func TestBase64IsAnotherJSON(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", `{"message": "eyJrZXkiOiJ2YWx1ZSJ9Cg==", "number": 42}`)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	expected := `{"message":{"key":"value"},"number":42}`
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Errorf("Expected: %s, Got: %s", expected, actual)
	}
}

// TestFileInputArgument tests reading JSON from a file argument
func TestFileInputArgument(t *testing.T) {
	// Create a temporary test file
	testJSON := `{
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

	// Write test file
	tmpFile := "test_input.json"
	err := os.WriteFile(tmpFile, []byte(testJSON), testFilePerms)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", tmpFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Parse the output to verify it's valid JSON and has expected decodings
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
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
}

// TestStdinRedirect tests reading JSON from stdin using file redirection
func TestStdinRedirect(t *testing.T) {
	testJSON := `{"data": "SGVsbG8="}`

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Stdin = strings.NewReader(testJSON)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	expected := `{"data":"Hello"}`
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Errorf("Expected: %s, Got: %s", expected, actual)
	}
}

// TestEmptyJSONObject tests with empty JSON object
func TestEmptyJSONObject(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", "{}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	expected := `{}`
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Errorf("Expected: %s, Got: %s", expected, actual)
	}
}

// TestPipeInput tests input via pipe (echo | go run main.go)
func TestPipeInput(t *testing.T) {
	testJSON := `{"encoded": "SGVsbG8=", "number": 123}`

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	// Create go run command
	goCmd := exec.CommandContext(ctx, "go", "run", "main.go")
	goCmd.Stdin = strings.NewReader(testJSON)

	// Run go command and get output
	output, err := goCmd.Output()
	if err != nil {
		t.Fatalf("Go command failed: %v", err)
	}

	// Verify the output
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if result["encoded"] != "Hello" {
		t.Errorf("Expected encoded field to be 'Hello', got '%v'", result["encoded"])
	}
	if result["number"].(float64) != 123 {
		t.Errorf("Expected number field to be 123, got '%v'", result["number"])
	}
}

// TestInvalidJSON tests error handling for invalid JSON
func TestInvalidJSON(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", `{"invalid": json}`)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected command to fail with invalid JSON")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Error parsing JSON") {
		t.Errorf("Expected error message about parsing JSON, got: %s", stderrOutput)
	}
}

// TestNonExistentFile tests error handling for non-existent file
func TestNonExistentFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", "nonexistent.json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected command to fail with non-existent file")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Error reading input") {
		t.Errorf("Expected error message about reading input, got: %s", stderrOutput)
	}
}

// TestComplexNestedJSON tests complex nested structures with mixed Base64 and regular data
func TestComplexNestedJSON(t *testing.T) {
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
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", complexJSON)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
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
}

// TestBase64EdgeCases tests various edge cases for Base64 detection and decoding
func TestBase64EdgeCases(t *testing.T) {
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
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", edgeCasesJSON)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
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
}

// TestShowHelpOnNoInput tests that help is displayed when no input is provided
func TestShowHelpOnNoInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	// Test with empty stdin
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Stdin = strings.NewReader("")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "JSON Base64 Decoder") {
		t.Error("Expected help message to be displayed when no input is provided")
	}
	if !strings.Contains(outputStr, "USAGE:") {
		t.Error("Expected help message to contain USAGE section")
	}
	if !strings.Contains(outputStr, "jbdecoder [OPTIONS] [INPUT]") {
		t.Error("Expected help message to contain usage pattern")
	}
}

// TestShowHelpOnEmptyInput tests that help is displayed when empty input is provided
func TestShowHelpOnEmptyInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	// Test with only whitespace stdin
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Stdin = strings.NewReader("   \n  \t  ")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "JSON Base64 Decoder") {
		t.Error("Expected help message to be displayed when empty input is provided")
	}
	if !strings.Contains(outputStr, "USAGE:") {
		t.Error("Expected help message to contain USAGE section")
	}
}

// TestBase64BinaryData tests that Base64 containing binary data is not corrupted
func TestBase64BinaryData(t *testing.T) {
	// Create a Base64 string that contains binary data (not valid UTF-8)
	binaryData := []byte{0xFF, 0xFE, 0x00, 0x01, 0x80, 0x81} // Invalid UTF-8 bytes
	binaryBase64 := base64.StdEncoding.EncodeToString(binaryData)

	testJSON := fmt.Sprintf(`{"binaryField": "%s", "textField": "SGVsbG8="}`, binaryBase64)

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go", testJSON)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// The binary Base64 should remain unchanged (not decoded)
	if result["binaryField"] != binaryBase64 {
		t.Errorf("Expected binaryField to remain as Base64 '%s', got '%v'", binaryBase64, result["binaryField"])
	}

	// The text Base64 should be decoded normally
	if result["textField"] != "Hello" {
		t.Errorf("Expected textField to be decoded to 'Hello', got '%v'", result["textField"])
	}
}
