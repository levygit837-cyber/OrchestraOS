package apperrors

import "fmt"

type Code string

const (
	CodeInvalidInput Code = "invalid_input"
	CodeValidation   Code = "validation"
	CodePersistence  Code = "persistence"
	CodeNotFound     Code = "not_found"
	CodeRuntime      Code = "runtime"
	CodeInternal     Code = "internal"
)

type Error struct {
	Code Code
	Op   string
	Err  error
}

func (e *Error) Error() string {
	if e.Op == "" {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Wrap(code Code, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Code: code, Op: op, Err: err}
}

func New(code Code, op, message string) error {
	return &Error{Code: code, Op: op, Err: fmt.Errorf("%s", message)}
}
