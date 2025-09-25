package errors

// ArgumentError represents an error related to command-line arguments
type ArgumentError struct {
	message string
}

// NewArgumentError creates a new ArgumentError with the given message
func NewArgumentError(message string) ArgumentError {
	return ArgumentError{message: message}
}

// Error implements the error interface for ArgumentError
func (e ArgumentError) Error() string {
	return e.message
}

// Global error variables for common argument errors
var (
	ErrEmptyInput      = NewArgumentError("empty input provided")
	ErrNoInputProvided = NewArgumentError("no input provided")
)
