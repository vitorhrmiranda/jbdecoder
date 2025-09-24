# JSON Base64 Decoder

A command-line utility written in Go that recursively decodes Base64 encoded strings in JSON data.

## Features

- **Recursive Processing**: Traverses nested JSON objects and arrays
- **Safe Decoding**: Only decodes valid Base64 strings, leaves other data unchanged
- **Multiple Input Methods**: Supports stdin, file input, and direct JSON arguments
- **Error Handling**: Clear error messages for malformed JSON or file issues
- **Help Documentation**: Built-in help with `-h` or `--help` flags

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd decode_grpc_response

# Build the binary
go build -o jbdecoder main.go

# Or run directly
go run main.go [arguments]
```

## Usage

### Command Line Options

```bash
jbdecoder [OPTIONS] [INPUT]
```

**Options:**
- `-h, --help`: Show help message and exit

### Input Methods

#### 1. Direct JSON String
```bash
go run main.go '{"message": "SGVsbG8gV29ybGQ=", "number": 42}'
# Output: {"message":"Hello World","number":42}
```

#### 2. File Input
```bash
go run main.go data.json
```

#### 3. Stdin (Pipe)
```bash
echo '{"data": "SGVsbG8="}' | go run main.go
# Output: {"data":"Hello"}
```

#### 4. File Redirection
```bash
go run main.go < input.json
```

### Examples

#### Simple Base64 Decoding
```bash
$ go run main.go '{"name": "Sm9obg==", "age": 30}'
{"name":"John","age":30}
```

#### Complex Nested JSON
```bash
$ go run main.go '{
  "user": {
    "name": "Sm9obiBEb2U=",
    "email": "am9obi5kb2VAZXhhbXBsZS5jb20="
  },
  "messages": ["SGVsbG8=", "V29ybGQ="],
  "count": 42,
  "active": true
}'
```

Output:
```json
{
  "user": {
    "name": "John Doe",
    "email": "john.doe@example.com"
  },
  "messages": ["Hello", "World"],
  "count": 42,
  "active": true
}
```

#### Mixed Data Types
```bash
$ go run main.go '{
  "valid_base64": "SGVsbG8gV29ybGQ=",
  "not_base64": "Hello@World",
  "number": 123,
  "boolean": true,
  "null_value": null,
  "array": ["VGVzdA==", "plain text", 456]
}'
```

Output:
```json
{
  "valid_base64": "Hello World",
  "not_base64": "Hello@World",
  "number": 123,
  "boolean": true,
  "null_value": null,
  "array": ["Test", "plain text", 456]
}
```

## How It Works

1. **Input Parsing**: Accepts JSON from various sources (file, stdin, argument)
2. **Validation**: Validates JSON syntax and Base64 format
3. **Recursive Processing**: Traverses all JSON structures (objects, arrays)
4. **Selective Decoding**: Only decodes strings that are valid Base64
5. **Output**: Returns processed JSON in compact format

## Base64 Detection

The tool identifies valid Base64 strings by:
- Checking if the string length is a multiple of 4
- Attempting to decode the string using Go's standard Base64 decoder
- Only strings that pass both checks are decoded

## Error Handling

The program handles errors gracefully:

- **Invalid JSON**: Returns clear error message and exits with code 1
- **File Not Found**: Shows file access error and exits with code 1
- **Too Many Arguments**: Displays usage error and exits with code 1

## Testing

Run the test suite:

```bash
go test -v
```

This runs comprehensive tests including:
- Direct JSON arguments
- File input/output
- Stdin processing
- Error handling scenarios
- Complex nested structures
- Base64 edge cases

## Development

### Code Quality

The project uses `golangci-lint` for code quality:

```bash
golangci-lint run
```

### Project Structure

```
├── main.go           # Main application code
├── main_test.go      # Comprehensive test suite
├── go.mod           # Go module definition
└── README.md        # This documentation
```