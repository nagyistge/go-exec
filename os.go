package exec

import (
	"os"
	"path/filepath"

	"github.com/codeship/go-concurrent"
	"github.com/codeship/go-osutils"
)

type osClientProvider struct {
	concurrent.Destroyable
	execOptions *OsExecOptions
}

func newOsExecutorReadFileManagerProvider(execOptions *OsExecOptions) *osClientProvider {
	return newOsClientProvider(execOptions)
}

func newOsExecutorWriteFileManagerProvider(execOptions *OsExecOptions) *osClientProvider {
	return newOsClientProvider(execOptions)
}

func newOsClientProvider(execOptions *OsExecOptions) *osClientProvider {
	return &osClientProvider{concurrent.NewDestroyable(nil), execOptions}
}

func (o *osClientProvider) NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error) {
	return o.NewTempDirClient()
}

func (o *osClientProvider) NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error) {
	return o.NewTempDirClient()
}

func (o *osClientProvider) NewTempDirClient() (Client, error) {
	tempDir, err := o.createTempDir()
	if err != nil {
		return nil, err
	}
	client := newOsClient(func() error { return o.removeTempDir(tempDir) }, tempDir)
	if err := o.AddChild(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (o *osClientProvider) createTempDir() (string, error) {
	value, err := o.Do(func() (interface{}, error) {
		if o.execOptions.TmpDir != "" {
			return osutils.NewTempSubDir(o.execOptions.TmpDir)
		}
		return osutils.NewTempDir()
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

// this is only called in thread-safe context
func (o *osClientProvider) removeTempDir(tempDir string) error {
	if err := o.validateIsDir(tempDir); err != nil {
		return err
	}
	return os.RemoveAll(tempDir)
}

// this is only called in thread-safe context
func (o *osClientProvider) validateIsDir(path string) error {
	exists, err := osutils.IsDirExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return ErrFileDoesNotExist
	}
	return nil
}

type osClient struct {
	concurrent.Destroyable
	dirPath string
}

func newOsAbsolutePathClient(absolutePath string) (*osClient, error) {
	if !filepath.IsAbs(absolutePath) {
		return nil, newValidationErrorNotAbsolutePath(absolutePath)
	}
	absolutePath, err := osutils.CleanPath(absolutePath)
	if err != nil {
		return nil, err
	}
	return newOsClient(func() error { return nil }, absolutePath), nil
}

func newOsClient(destroyCallback func() error, dirPath string) *osClient {
	return &osClient{concurrent.NewDestroyable(destroyCallback), dirPath}
}

func (o *osClient) DirName() string {
	return filepath.Base(o.DirPath())
}

func (o *osClient) DirPath() string {
	return o.dirPath
}

func (o *osClient) Execute(cmd *Cmd) func() error {
	if cmd.SubDir != "" {
		if err := o.validatePath(cmd.SubDir); err != nil {
			return func() error { return err }
		}
	}
	value, err := o.Do(func() (interface{}, error) {
		return osutils.Execute(o.osutilsCmd(cmd))
	})
	if err != nil {
		return func() error { return err }
	}
	return value.(func() error)
}

func (o *osClient) ExecutePiped(pipeCmdList *PipeCmdList) func() error {
	for _, pipeCmd := range pipeCmdList.PipeCmds {
		if pipeCmd.SubDir != "" {
			if err := o.validatePath(pipeCmd.SubDir); err != nil {
				return func() error { return err }
			}
		}
	}
	value, err := o.Do(func() (interface{}, error) {
		return osutils.ExecutePiped(o.osutilsPipeCmdList(pipeCmdList))
	})
	if err != nil {
		return func() error { return err }
	}
	return value.(func() error)
}

func (o *osClient) IsFileExists(path string) (bool, error) {
	if err := o.validatePath(path); err != nil {
		return false, err
	}
	value, err := o.Do(func() (interface{}, error) {
		return osutils.IsFileExists(o.absolutePath(path))
	})
	if err != nil {
		return false, err
	}
	return value.(bool), nil
}

func (o *osClient) Open(path string) (ReadFile, error) {
	if err := o.validatePath(path); err != nil {
		return nil, err
	}
	value, err := o.Do(func() (interface{}, error) {
		return osutils.Open(o.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*os.File), nil
}

func (o *osClient) Create(path string) (WriteFile, error) {
	if err := o.validatePath(path); err != nil {
		return nil, err
	}
	value, err := o.Do(func() (interface{}, error) {
		return osutils.Create(o.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*os.File), nil
}

func (o *osClient) MkdirAll(path string, perm os.FileMode) error {
	if err := o.validatePath(path); err != nil {
		return err
	}
	_, err := o.Do(func() (interface{}, error) {
		return nil, osutils.MkdirAll(o.absolutePath(path), perm)
	})
	return err
}

func (o *osClient) Rename(oldpath string, newpath string) error {
	if err := o.validatePath(oldpath); err != nil {
		return err
	}
	if err := o.validatePath(newpath); err != nil {
		return err
	}
	_, err := o.Do(func() (interface{}, error) {
		return nil, os.Rename(o.absolutePath(oldpath), o.absolutePath(newpath))
	})
	return err
}

func (o *osClient) Remove(path string) error {
	if err := o.validatePath(path); err != nil {
		return err
	}
	_, err := o.Do(func() (interface{}, error) {
		return nil, os.Remove(o.absolutePath(path))
	})
	return err
}

func (o *osClient) ListRegularFiles(path string) ([]string, error) {
	if err := o.validatePath(path); err != nil {
		return nil, err
	}
	value, err := o.Do(func() (interface{}, error) {
		files, err := osutils.ListRegularFiles(o.absolutePath(path))
		if err != nil {
			return nil, err
		}
		relFiles := make([]string, len(files))
		for i, file := range files {
			rel, err := filepath.Rel(o.dirPath, file)
			if err != nil {
				return nil, err
			}
			relFiles[i] = rel
		}
		return relFiles, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]string), nil
}

func (o *osClient) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (o *osClient) Match(pattern string, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

func (o *osClient) ToSlash(path string) string {
	return filepath.ToSlash(path)
}

func (o *osClient) Base(path string) string {
	return filepath.Base(path)
}

func (o *osClient) Dir(path string) string {
	return filepath.Dir(path)
}

func (o *osClient) PathSeparator() string {
	return string(os.PathSeparator)
}

func (o *osClient) NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error) {
	return o.newSubDirClient(path)
}

func (o *osClient) NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error) {
	return o.newSubDirClient(path)
}

func (o *osClient) NewSubDirClient(path string) (Client, error) {
	return o.newSubDirClient(path)
}

func (o *osClient) newSubDirClient(path string) (*osClient, error) {
	if err := o.validatePath(path); err != nil {
		return nil, err
	}
	exists, err := o.IsFileExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFileAlreadyExists
	}
	if err := osutils.Mkdir(o.absolutePath(path), 0755); err != nil {
		return nil, err
	}
	subDirClient := newOsClient(func() error { return o.removeDir(path) }, o.absolutePath(path))
	if err := o.AddChild(subDirClient); err != nil {
		return nil, err
	}
	return subDirClient, nil
}

func (o *osClient) removeDir(path string) error {
	if err := o.validateIsDir(path); err != nil {
		return err
	}
	return osutils.RemoveAll(o.absolutePath(path))
}

func (o *osClient) validateIsDir(path string) error {
	if err := o.validatePath(path); err != nil {
		return err
	}
	exists, err := osutils.IsDirExists(o.absolutePath(path))
	if err != nil {
		return err
	}
	if !exists {
		return ErrFileDoesNotExist
	}
	return nil
}

func (o *osClient) validatePath(path string) error {
	if filepath.IsAbs(path) {
		return ErrNotRelativePath
	}
	// TODO(pedge): EvalSymlinks fails if the file does not exist
	//path, err := filepath.EvalSymlinks(filepath.Clean(o.absolutePath(path)))
	//if err != nil {
	//return err
	//}
	//if !strings.HasPrefix(path, o.DirPath()) {
	//return ErrPathOutOfContext
	//}
	return nil
}

func (o *osClient) absolutePath(path string) string {
	return o.Join(o.dirPath, path)
}

func (o *osClient) osutilsCmd(cmd *Cmd) *osutils.Cmd {
	return &osutils.Cmd{
		Args:        cmd.Args,
		AbsoluteDir: o.absolutePath(cmd.SubDir),
		Env:         cmd.Env,
		Stdin:       cmd.Stdin,
		Stdout:      cmd.Stdout,
		Stderr:      cmd.Stderr,
	}
}

func (o *osClient) osutilsPipeCmdList(pipeCmdList *PipeCmdList) *osutils.PipeCmdList {
	pipeCmds := make([]*osutils.PipeCmd, len(pipeCmdList.PipeCmds))
	for i, pipeCmd := range pipeCmdList.PipeCmds {
		pipeCmds[i] = &osutils.PipeCmd{
			Args:        pipeCmd.Args,
			AbsoluteDir: o.absolutePath(pipeCmd.SubDir),
			Env:         pipeCmd.Env,
		}
	}
	return &osutils.PipeCmdList{
		PipeCmds: pipeCmds,
		Stdin:    pipeCmdList.Stdin,
		Stdout:   pipeCmdList.Stdout,
		Stderr:   pipeCmdList.Stderr,
	}
}
