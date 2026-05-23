package apperrors

import "fmt"

type Kind int

const (
	KindValidation Kind = iota
	KindNotFound
	KindConflict
	KindInternal
)

type Error struct {
	Kind    Kind
	Op      string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func New(kind Kind, op, msg string) *Error {
	return &Error{Kind: kind, Op: op, Message: msg}
}

func Wrap(kind Kind, op string, err error) *Error {
	return &Error{Kind: kind, Op: op, Message: err.Error(), Err: err}
}

func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Kind == KindNotFound
	}
	return false
}
