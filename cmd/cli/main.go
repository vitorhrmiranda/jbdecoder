package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/vitorhrmiranda/jbdecoder/internal/decoder"
	errs "github.com/vitorhrmiranda/jbdecoder/internal/errors"
)

//go:embed help.md
var helpTemplate string

const (
	Zero = iota
	One
)

// showUsage displays the help message
func showUsage() {
	tmpl, err := template.New("help").Parse(helpTemplate)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing help template: %v\n", err)
		return
	}

	const programName = "jbdecoder"
	if err := tmpl.Execute(os.Stdout, programName); err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"Error executing help template: %v\n", err)
	}
}

// isStdinEmpty checks if stdin has no data available
func isStdinEmpty() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return stat.Mode()&os.ModeCharDevice != Zero
}

// readFromStdin reads and validates input from standard input
func readFromStdin() ([]byte, error) {
	if isStdinEmpty() {
		return nil, errs.ErrNoInputProvided
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(string(data))) == Zero {
		return nil, errs.ErrEmptyInput
	}

	return data, nil
}

// processArgument handles a single command-line argument (JSON string or file)
func processArgument(arg string) ([]byte, error) {
	arg = strings.TrimSpace(arg)

	// Check if the argument looks like a JSON string (starts with { or [)
	if strings.HasPrefix(arg, "{") || strings.HasPrefix(arg, "[") {
		return []byte(arg), nil
	}

	// Otherwise, treat it as a filename
	file, err := os.Open(arg)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", arg, err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// getJSONInput reads JSON input from various sources
func getJSONInput() ([]byte, error) {
	args := flag.Args()

	switch len(args) {
	case Zero:
		return readFromStdin()
	case One:
		return processArgument(args[Zero])
	default:
		return nil, errors.New("too many arguments provided")
	}
}

func main() {
	help := flag.Bool("h", false, "Show help message")
	flag.BoolVar(help, "help", false, "Show help message")
	flag.Usage = showUsage
	flag.Parse()

	if *help {
		showUsage()
		return
	}

	jsonData, err := getJSONInput()
	if err != nil {
		// Check if error is an ArgumentError - show help instead of error
		var argErr errs.ArgumentError
		if errors.As(err, &argErr) {
			showUsage()
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(One)
	}

	var data any
	if parseErr := json.Unmarshal(jsonData, &data); parseErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", parseErr)
		os.Exit(One)
	}

	processedData := decoder.DecodeBase64Fields(data)

	output, err := json.Marshal(processedData)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error generating output JSON: %v\n", err)
		os.Exit(One)
	}

	_, _ = fmt.Println(string(output))
}
