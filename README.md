[![Codeship Status](http://img.shields.io/codeship/34b974b0-6dfa-0132-51b4-66f2bf861e14/master.svg?style=flat-square)](https://codeship.com/projects/57533)
[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/peter-edge/exec)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/peter-edge/exec/blob/master/LICENSE)

## Installation
```bash
go get -u gopkg.in/peter-edge/exec.v1
```

## Import
```go
import (
    "gopkg.in/peter-edge/exec.v1"
)
```


```go
var (
	ErrAlreadyDestroyed = errors.New("exec: already destroyed")
	ErrFileDoesNotExist = errors.New("exec: file does not exist")
	ErrNotRelativePath  = errors.New("exec: not relative path")
	ErrPathOutOfContext = errors.New("exec: path out of context")
	ErrArgsEmpty        = errors.New("exec: args empty")
)
```

#### type Client

```go
type Client interface {
	DirContext
	Destroyable
	Execute(cmd *Cmd) func() error
	IsFileExists(path string) (bool, error)
	Open(path string) (ReadFile, error)
	ListRegularFiles(path string) ([]string, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	PathSeparator() string
	NewSubDirClient(string) (Client, error)
}
```


#### type ClientProvider

```go
type ClientProvider interface {
	Destroyable
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


#### type File

```go
type File interface {
	Stat() (os.FileInfo, error)
	Close() error
}
```


#### type FileManager

```go
type FileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	Open(path string) (ReadFile, error)
	ListRegularFiles(path string) ([]string, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	PathSeparator() string
}
```

All paths must be relative

#### type ReadFile

```go
type ReadFile interface {
	File
	io.Reader
	Readdirnames(n int) ([]string, error)
}
```
