package exec

import (
	"io"
	"os"
)

type ExternalExecOptions struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

func NewExternalExecutorReadFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorReadFileManagerProvider, error) {
	return NewExternalClientProvider(externalExecOptions)
}

func NewExternalExecutorWriteFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorWriteFileManagerProvider, error) {
	return NewExternalClientProvider(externalExecOptions)
}

func NewExternalClientProvider(externalExecOptions *ExternalExecOptions) (ClientProvider, error) {
	return newExternalClientProvider(externalExecOptions)
}

type ExecOptions interface {
	Type() ExecType
}

type OsExecOptions struct{}

func (this *OsExecOptions) Type() ExecType {
	return ExecTypeOs
}

func NewExecutorReadFileManagerProvider(execOptions ExecOptions) (ExecutorReadFileManagerProvider, error) {
	return NewClientProvider(execOptions)
}

func NewExecutorWriteFileManagerProvider(execOptions ExecOptions) (ExecutorWriteFileManagerProvider, error) {
	return NewClientProvider(execOptions)
}

func NewClientProvider(execOptions ExecOptions) (ClientProvider, error) {
	return newClientProvider(execOptions)
}

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

type PipeCmd struct {
	// Includes path
	Args []string

	// can be empty
	// must be relative
	SubDir string

	// can be nil or empty
	Env []string
}

type PipeCmdList struct {
	PipeCmds []*PipeCmd

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

type WriteFile interface {
	File
	io.Writer
	Chmod(mode os.FileMode) error
}

type Destroyable interface {
	Destroy() error
	Do(func() (interface{}, error)) (interface{}, error)
	AddChild(Destroyable) error
}

type DirContext interface {
	DirPath() string
}

type Executor interface {
	DirContext
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
}

// All paths must be relative
type ReadFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	ListFileInfosShallow(path string) ([]os.FileInfo, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Dir(path string) string
	PathSeparator() string
	Open(path string) (ReadFile, error)
}

type ExecutorReadFileManager interface {
	Destroyable
	ReadFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
}

type ExecutorReadFileManagerProvider interface {
	Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
}

// All paths must be relative
type WriteFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	ListFileInfosShallow(path string) ([]os.FileInfo, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Dir(path string) string
	PathSeparator() string
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
}

type ExecutorWriteFileManager interface {
	Destroyable
	WriteFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
}

type ExecutorWriteFileManagerProvider interface {
	Destroyable
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
}

type ReadWriteFileManager interface {
	DirContext
	IsFileExists(path string) (bool, error)
	ListRegularFiles(path string) ([]string, error)
	ListFileInfosShallow(path string) ([]os.FileInfo, error)
	Join(elem ...string) string
	Match(pattern string, path string) (bool, error)
	ToSlash(path string) string
	Dir(path string) string
	PathSeparator() string
	Open(path string) (ReadFile, error)
	Create(name string) (WriteFile, error)
	MkdirAll(path string, perm os.FileMode) error
}

type Client interface {
	Destroyable
	ReadWriteFileManager
	Execute(cmd *Cmd) func() error
	ExecutePiped(pipeCmdList *PipeCmdList) func() error
	NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error)
	NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error)
	NewSubDirClient(path string) (Client, error)
}

type ClientProvider interface {
	Destroyable
	NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error)
	NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error)
	NewTempDirClient() (Client, error)
}
