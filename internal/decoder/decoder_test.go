package decoder_test

import (
	"encoding/json"
	"testing"

	"github.com/vitorhrmiranda/jbdecoder/internal/decoder"
)

const example = `{
	"data": "eyJtZXNzYWdlIjoiZGlzdGFuY2UifQo="
}`

func Test_DecodeBase64Fields(t *testing.T) {
	var data any
	_ = json.Unmarshal([]byte(example), &data)
	decoded := decoder.DecodeBase64Fields(data)
	jdecoded, _ := json.Marshal(decoded)

	// Parse expected result to compare content rather than string representation
	expected := `{"data":{"message":"distance"}}`
	if expected != string(jdecoded) {
		t.Errorf("Expected: %s, Got: %s", expected, decoded)
	}
}
