package exec

import (
	"errors"
	"fmt"
)

var (
	ErrAlreadyDestroyed    = errors.New("exec: already destroyed")
	ErrFileDoesNotExist    = errors.New("exec: file does not exist")
	ErrNotRelativePath     = errors.New("exec: not relative path")
	ErrNotAbsolutePath     = errors.New("exec: not absolute path")
	ErrPathOutOfContext    = errors.New("exec: path out of context")
	ErrArgsEmpty           = errors.New("exec: args empty")
	ErrFileAlreadyExists   = errors.New("exec: file already exists")
	ErrNotMultipleCommands = errors.New("exec: not multiple commands")
	ErrNotADirectory       = errors.New("exec: not a directory")

	ValidationErrorTypeNotAbsolutePath ValidationErrorType = "NotAbsolutePath"
	ValidationErrorTypeUnknownExecType ValidationErrorType = "UnknownExecType"
)

type ValidationErrorType string

type ValidationError interface {
	error
	Type() ValidationErrorType
}
type validationError struct {
	errorType ValidationErrorType
	tags      map[string]string
}

func newValidationError(errorType ValidationErrorType, tags map[string]string) *validationError {
	if tags == nil {
		tags = make(map[string]string)
	}
	return &validationError{errorType, tags}
}

func (v *validationError) Error() string {
	return fmt.Sprintf("%v %v", v.errorType, v.tags)
}

func (v *validationError) Type() ValidationErrorType {
	return v.errorType
}

func newValidationErrorUnknownExecType(execType string) ValidationError {
	return newValidationError(ValidationErrorTypeUnknownExecType, map[string]string{"execType": execType})
}

func newValidationErrorNotAbsolutePath(path string) ValidationError {
	return newValidationError(ValidationErrorTypeNotAbsolutePath, map[string]string{"path": path})
}

func newInternalError(validationError ValidationError) error {
	return errors.New(validationError.Error())
}
