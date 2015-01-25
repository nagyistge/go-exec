package exec

import "errors"

var (
	ErrAlreadyDestroyed  = errors.New("exec: already destroyed")
	ErrFileDoesNotExist  = errors.New("exec: file does not exist")
	ErrNotRelativePath   = errors.New("exec: not relative path")
	ErrPathOutOfContext  = errors.New("exec: path out of context")
	ErrArgsEmpty         = errors.New("exec: args empty")
	ErrFileAlreadyExists = errors.New("exec: file already exists")
)
