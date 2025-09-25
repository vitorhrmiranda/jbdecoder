# JSON Base64 Decoder

A command-line utility to recursively decode Base64 encoded strings in JSON.

## USAGE:
  {{.}} [INPUT]

## INPUT METHODS:
  # Read from stdin (pipe)
  echo '{"data": "SGVsbG8="}' | {{.}}

  # Read from file (redirect)
  {{.}} < input.json

  # Read from file (argument)
  {{.}} input.json

  # Direct JSON string argument
  {{.}} '{"message": "SGVsbG8gV29ybGQ="}'

## DESCRIPTION:
  This tool recursively traverses JSON data and decodes any string fields
  that contain valid Base64 encoded data. Other data types (numbers,
  booleans, arrays, objects) are preserved unchanged.

  Only strings that are valid Base64 will be decoded. Invalid Base64
  strings are left unchanged.

## OPTIONS:
  -h, --help    Show this help message and exit

## EXAMPLES:
  # Decode Base64 strings in a JSON file
  {{.}} data.json

  # Decode from stdin
  cat data.json | {{.}}

  # Decode a simple JSON string
  {{.}} '{"name": "Sm9obg==", "age": 30}'

  # Handle complex nested JSON
  {{.}} '{"user": {"token": "dG9rZW4="}, "items": ["aXRlbTE="]}'
