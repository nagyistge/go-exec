package exec

import (
	"os"
	"path/filepath"

	"github.com/peter-edge/go-concurrent"
	"github.com/peter-edge/go-osutils"
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

func (this *osClientProvider) NewTempDirExecutorReadFileManager() (ExecutorReadFileManager, error) {
	return this.NewTempDirClient()
}

func (this *osClientProvider) NewTempDirExecutorWriteFileManager() (ExecutorWriteFileManager, error) {
	return this.NewTempDirClient()
}

func (this *osClientProvider) NewTempDirClient() (Client, error) {
	tempDir, err := this.createTempDir()
	if err != nil {
		return nil, err
	}
	client := newOsClient(func() error { return this.removeTempDir(tempDir) }, tempDir)
	if err := this.AddChild(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (this *osClientProvider) createTempDir() (string, error) {
	value, err := this.Do(func() (interface{}, error) {
		if this.execOptions.TmpDir != "" {
			return osutils.NewTempSubDir(this.execOptions.TmpDir)
		} else {
			return osutils.NewTempDir()
		}
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

// this is only called in thread-safe context
func (this *osClientProvider) removeTempDir(tempDir string) error {
	if err := this.validateIsDir(tempDir); err != nil {
		return err
	}
	return os.RemoveAll(tempDir)
}

// this is only called in thread-safe context
func (this *osClientProvider) validateIsDir(path string) error {
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

func (this *osClient) DirName() string {
	return filepath.Base(this.DirPath())
}

func (this *osClient) DirPath() string {
	return this.dirPath
}

func (this *osClient) Execute(cmd *Cmd) func() error {
	if cmd.SubDir != "" {
		if err := this.validatePath(cmd.SubDir); err != nil {
			return func() error { return err }
		}
	}
	value, err := this.Do(func() (interface{}, error) {
		return osutils.Execute(this.osutilsCmd(cmd))
	})
	if err != nil {
		return func() error { return err }
	}
	return value.(func() error)
}

func (this *osClient) ExecutePiped(pipeCmdList *PipeCmdList) func() error {
	for _, pipeCmd := range pipeCmdList.PipeCmds {
		if pipeCmd.SubDir != "" {
			if err := this.validatePath(pipeCmd.SubDir); err != nil {
				return func() error { return err }
			}
		}
	}
	value, err := this.Do(func() (interface{}, error) {
		return osutils.ExecutePiped(this.osutilsPipeCmdList(pipeCmdList))
	})
	if err != nil {
		return func() error { return err }
	}
	return value.(func() error)
}

func (this *osClient) IsFileExists(path string) (bool, error) {
	if err := this.validatePath(path); err != nil {
		return false, err
	}
	value, err := this.Do(func() (interface{}, error) {
		return osutils.IsFileExists(this.absolutePath(path))
	})
	if err != nil {
		return false, err
	}
	return value.(bool), nil
}

func (this *osClient) Open(path string) (ReadFile, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		return osutils.Open(this.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*os.File), nil
}

func (this *osClient) Create(path string) (WriteFile, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		return osutils.Create(this.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*os.File), nil
}

func (this *osClient) MkdirAll(path string, perm os.FileMode) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	_, err := this.Do(func() (interface{}, error) {
		return nil, osutils.MkdirAll(this.absolutePath(path), perm)
	})
	return err
}

func (this *osClient) Rename(oldpath string, newpath string) error {
	if err := this.validatePath(oldpath); err != nil {
		return err
	}
	if err := this.validatePath(newpath); err != nil {
		return err
	}
	_, err := this.Do(func() (interface{}, error) {
		return nil, os.Rename(this.absolutePath(oldpath), this.absolutePath(newpath))
	})
	return err
}

func (this *osClient) Remove(path string) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	_, err := this.Do(func() (interface{}, error) {
		return nil, os.Remove(this.absolutePath(path))
	})
	return err
}

func (this *osClient) ListRegularFiles(path string) ([]string, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		files, err := osutils.ListRegularFiles(this.absolutePath(path))
		if err != nil {
			return nil, err
		}
		relFiles := make([]string, len(files))
		for i, file := range files {
			rel, err := filepath.Rel(this.dirPath, file)
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

func (this *osClient) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (this *osClient) Match(pattern string, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

func (this *osClient) ToSlash(path string) string {
	return filepath.ToSlash(path)
}

func (this *osClient) Base(path string) string {
	return filepath.Base(path)
}

func (this *osClient) Dir(path string) string {
	return filepath.Dir(path)
}

func (this *osClient) PathSeparator() string {
	return string(os.PathSeparator)
}

func (this *osClient) NewSubDirExecutorReadFileManager(path string) (ExecutorReadFileManager, error) {
	return this.newSubDirClient(path)
}

func (this *osClient) NewSubDirExecutorWriteFileManager(path string) (ExecutorWriteFileManager, error) {
	return this.newSubDirClient(path)
}

func (this *osClient) NewSubDirClient(path string) (Client, error) {
	return this.newSubDirClient(path)
}

func (this *osClient) newSubDirClient(path string) (*osClient, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	exists, err := this.IsFileExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFileAlreadyExists
	}
	if err := osutils.Mkdir(this.absolutePath(path), 0755); err != nil {
		return nil, err
	}
	subDirClient := newOsClient(func() error { return this.removeDir(path) }, this.absolutePath(path))
	if err := this.AddChild(subDirClient); err != nil {
		return nil, err
	}
	return subDirClient, nil
}

func (this *osClient) removeDir(path string) error {
	if err := this.validateIsDir(path); err != nil {
		return err
	}
	return osutils.RemoveAll(this.absolutePath(path))
}

func (this *osClient) validateIsDir(path string) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	exists, err := osutils.IsDirExists(this.absolutePath(path))
	if err != nil {
		return err
	}
	if !exists {
		return ErrFileDoesNotExist
	}
	return nil
}

func (this *osClient) validatePath(path string) error {
	if filepath.IsAbs(path) {
		return ErrNotRelativePath
	}
	// TODO(pedge): EvalSymlinks fails if the file does not exist
	//path, err := filepath.EvalSymlinks(filepath.Clean(this.absolutePath(path)))
	//if err != nil {
	//return err
	//}
	//if !strings.HasPrefix(path, this.DirPath()) {
	//return ErrPathOutOfContext
	//}
	return nil
}

func (this *osClient) absolutePath(path string) string {
	return this.Join(this.dirPath, path)
}

func (this *osClient) osutilsCmd(cmd *Cmd) *osutils.Cmd {
	return &osutils.Cmd{
		Args:        cmd.Args,
		AbsoluteDir: this.absolutePath(cmd.SubDir),
		Env:         cmd.Env,
		Stdin:       cmd.Stdin,
		Stdout:      cmd.Stdout,
		Stderr:      cmd.Stderr,
	}
}

func (this *osClient) osutilsPipeCmdList(pipeCmdList *PipeCmdList) *osutils.PipeCmdList {
	pipeCmds := make([]*osutils.PipeCmd, len(pipeCmdList.PipeCmds))
	for i, pipeCmd := range pipeCmdList.PipeCmds {
		pipeCmds[i] = &osutils.PipeCmd{
			Args:        pipeCmd.Args,
			AbsoluteDir: this.absolutePath(pipeCmd.SubDir),
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
