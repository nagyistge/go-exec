[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/peter-edge/go-exec)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/peter-edge/go-exec/blob/master/LICENSE)

## Installation

```bash
go get -u github.com/peter-edge/go-exec
```

## Import

```go
import (
    "github.com/peter-edge/go-exec"
)
```

## Usage

```go
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
```

#### func  AllExecTypes

```go
func AllExecTypes() []ExecType
```

#### func  UnknownExecType

```go
func UnknownExecType(unknownExecType interface{}) error
```

#### func  ValidateExecOptions

```go
func ValidateExecOptions(execOptions ExecOptions) error
```

#### type Client

```go
type Client interface {
	concurrent.Destroyable
	ReadWriteFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
	NewSubDirClient(path string) (Client, error)
}
```


#### type ClientProvider

```go
type ClientProvider interface {
	concurrent.Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
	NewTempDirClient() (Client, error)
}
```


#### func  NewClientProvider

```go
func NewClientProvider(execOptions ExecOptions) (ClientProvider, error)
```

#### func  NewExternalClientProvider

```go
func NewExternalClientProvider(externalExecOptions *ExternalExecOptions) (ClientProvider, error)
```

#### type Cmd

```go
type Cmd struct {
	// Includes path
	Args []string

	// can be empty
	// must be relative
	SubDir string

	// can be nil or empty
	Env []string

	// Can be nil
	Stdin io.Reader
	// Can be nil
	Stdout io.Writer
	// Can be nil
	Stderr io.Writer
}
```


#### type DirContext

```go
type DirContext interface {
	DirPath() string
}
```


#### type ExecOptions

```go
type ExecOptions interface {
	Type() ExecType
}
```


#### func  ConvertExternalExecOptions

```go
func ConvertExternalExecOptions(externalExecOptions *ExternalExecOptions) (ExecOptions, error)
```

#### type ExecType

```go
type ExecType uint
```


```go
var (
	ExecTypeOs ExecType = 0
)
```

#### func  ExecTypeOf

```go
func ExecTypeOf(s string) (ExecType, error)
```

#### func (ExecType) String

```go
func (this ExecType) String() string
```

#### type Executor

```go
type Executor interface {
	DirContext
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
}
```


#### func  NewOsExecutor

```go
func NewOsExecutor(absolutePath string) (Executor, error)
```

#### type ExecutorReadFileManager

```go
type ExecutorReadFileManager interface {
	concurrent.Destroyable
	ReadFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
}
```


#### type ExecutorReadFileManagerProvider

```go
type ExecutorReadFileManagerProvider interface {
	concurrent.Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
}
```


#### func  NewExecutorReadFileManagerProvider

```go
func NewExecutorReadFileManagerProvider(execOptions ExecOptions) (ExecutorReadFileManagerProvider, error)
```

#### func  NewExternalExecutorReadFileManagerProvider

```go
func NewExternalExecutorReadFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorReadFileManagerProvider, error)
```

#### type ExecutorWriteFileManager

```go
type ExecutorWriteFileManager interface {
	concurrent.Destroyable
	WriteFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
}
```


#### type ExecutorWriteFileManagerProvider

```go
type ExecutorWriteFileManagerProvider interface {
	concurrent.Destroyable
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
}
```


#### func  NewExecutorWriteFileManagerProvider

```go
func NewExecutorWriteFileManagerProvider(execOptions ExecOptions) (ExecutorWriteFileManagerProvider, error)
```

#### func  NewExternalExecutorWriteFileManagerProvider

```go
func NewExternalExecutorWriteFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorWriteFileManagerProvider, error)
```

#### type ExternalExecOptions

```go
type ExternalExecOptions struct {
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	TmpDir string `json:"tmp_dir,omitempty" yaml:tmp_dir,omitempty"`
}
```


#### type File

```go
type File interface {
	Stat() (os.FileInfo, error)
	Close() error
}
```


#### type OsExecOptions

```go
type OsExecOptions struct {
	TmpDir string
}
```


#### func (*OsExecOptions) Type

```go
func (this *OsExecOptions) Type() ExecType
```

#### type PipeCmd

```go
type PipeCmd struct {
	// Includes path
	Args []string

	// can be empty
	// must be relative
	SubDir string

	// can be nil or empty
	Env []string
}
```


#### type PipeCmdList

```go
type PipeCmdList struct {
	PipeCmds []*PipeCmd

	// Can be nil
	Stdin io.Reader
	// Can be nil
	Stdout io.Writer
	// Can be nil
	Stderr io.Writer
}
```


#### type ReadFile

```go
type ReadFile interface {
	File
	io.Reader
	Readdir(n int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
}
```


#### type ReadFileManager

```go
type ReadFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Base(path string) string
	Dir(path string) string
	PathSeparator() string
	Open(path string) (ReadFile, error)
}
```

All paths must be relative

#### type ReadWriteFileManager

```go
type ReadWriteFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Base(path string) string
	Dir(path string) string
	PathSeparator() string
	Open(path string) (ReadFile, error)
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
	Rename(oldpath string, newpath string) error
	Remove(path string) error
}
```


#### func  NewOsReadWriteFileManager

```go
func NewOsReadWriteFileManager(absolutePath string) (ReadWriteFileManager, error)
```

#### type ValidationError

```go
type ValidationError interface {
	error
	Type() ValidationErrorType
}
```


#### type ValidationErrorType

```go
type ValidationErrorType string
```


#### type WriteFile

```go
type WriteFile interface {
	File
	io.Writer
	Chmod(mode os.FileMode) error
}
```


#### type WriteFileManager

```go
type WriteFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Base(path string) string
	Dir(path string) string
	PathSeparator() string
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
	Rename(oldpath string, newpath string) error
	Remove(path string) error
}
```

All paths must be relative
