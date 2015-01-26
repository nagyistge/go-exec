[![Codeship Status](http://img.shields.io/codeship/34b974b0-6dfa-0132-51b4-66f2bf861e14/master.svg?style=flat-square)](https://codeship.com/projects/59077)
[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/peter-edge/exec)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/peter-edge/exec/blob/master/LICENSE)

## Installation
```bash
go get -u github.com/peter-edge/exec
```

## Import
```go
import (
    "github.com/peter-edge/exec"
)
```

## Usage

```go
var (
	ErrAlreadyDestroyed  = errors.New("exec: already destroyed")
	ErrFileDoesNotExist  = errors.New("exec: file does not exist")
	ErrNotRelativePath   = errors.New("exec: not relative path")
	ErrPathOutOfContext  = errors.New("exec: path out of context")
	ErrArgsEmpty         = errors.New("exec: args empty")
	ErrFileAlreadyExists = errors.New("exec: file already exists")
)
```

#### type Client

```go
type Client interface {
	Destroyable
	ReadWriteFileManager
	Execute(cmd *Cmd) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
	NewSubDirClient(path string) (Client, error)
}
```


#### type ClientProvider

```go
type ClientProvider interface {
	Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
	NewTempDirClient() (Client, error)
}
```


#### type Cmd

```go
type Cmd struct {
	// Includes path
	Args []string

	// Can be nil
	Stdin io.Reader
	// Can be nil
	Stdout io.Writer
	// Can be nil
	Stderr io.Writer
}
```


#### type Destroyable

```go
type Destroyable interface {
	Destroy() error
	Do(func() (interface{}, error)) (interface{}, error)
	AddChild(Destroyable) error
}
```


#### func  NewDestroyable

```go
func NewDestroyable(destroyCallback func() error) Destroyable
```

#### type DirContext

```go
type DirContext interface {
	DirPath() string
}
```


#### type Executor

```go
type Executor interface {
	DirContext
	Execute(cmd *Cmd) func() error
}
```


#### type ExecutorReadFileManager

```go
type ExecutorReadFileManager interface {
	Destroyable
	ReadFileManager
	Execute(cmd *Cmd) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
}
```


#### type ExecutorReadFileManagerProvider

```go
type ExecutorReadFileManagerProvider interface {
	Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
}
```


#### type ExecutorWriteFileManager

```go
type ExecutorWriteFileManager interface {
	Destroyable
	WriteFileManager
	Execute(cmd *Cmd) func() error
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
}
```


#### type ExecutorWriteFileManagerProvider

```go
type ExecutorWriteFileManagerProvider interface {
	Destroyable
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
}
```


#### type File

```go
type File interface {
	Stat() (os.FileInfo, error)
	Close() error
}
```


#### type ReadFile

```go
type ReadFile interface {
	File
	io.Reader
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
	Dir(path string) string
	PathSeparator() string
	Open(path string) (ReadFile, error)
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
}
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
	Dir(path string) string
	PathSeparator() string
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
}
```

All paths must be relative
