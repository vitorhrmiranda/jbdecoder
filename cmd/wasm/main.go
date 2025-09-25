//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/vitorhrmiranda/jbdecoder/internal/decoder"
)

// decodeJSON is the main function exposed to JavaScript
func decodeJSON(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return map[string]any{
			"error": "Expected exactly one argument (JSON string)",
		}
	}

	jsonStr := args[0].String()

	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return map[string]any{
			"error": "Invalid JSON: " + err.Error(),
		}
	}

	// Process the JSON data to decode Base64 fields using the decoder module
	processedData := decoder.DecodeBase64Fields(data)

	// Convert back to JSON
	output, err := json.Marshal(processedData)
	if err != nil {
		return map[string]any{
			"error": "Error generating output JSON: " + err.Error(),
		}
	}

	return map[string]any{
		"result": string(output),
	}
}

// main function registers the WebAssembly functions
func main() {
	c := make(chan struct{})

	// Register the decodeJSON function to be called from JavaScript
	js.Global().Set("decodeJSON", js.FuncOf(decodeJSON))

	// Signal that WASM is ready
	js.Global().Set("wasmReady", js.ValueOf(true))

	// Keep the program running
	<-c
}
