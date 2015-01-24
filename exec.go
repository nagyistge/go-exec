package exec

import (
	"io"
	"os"
)

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

type File interface {
	Stat() (os.FileInfo, error)
	Close() error
}

type ReadFile interface {
	File
	io.Reader
	Readdirnames(n int) ([]string, error)
}

type DirContext interface {
	DirPath() string
}

type Executor interface {
	DirContext
	Execute(cmd *Cmd) func() error
}

// All paths must be relative
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

type Destroyable interface {
	Destroy() error
	Do(func() (interface{}, error)) (interface{}, error)
	AddChild(Destroyable) error
}

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

type ClientProvider interface {
	Destroyable
	NewTempDirClient() (Client, error)
}
