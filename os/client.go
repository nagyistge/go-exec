package os

import (
	stdos "os"
	stdosexec "os/exec"
	"path/filepath"

	"gopkg.in/peter-edge/exec.v1"
)

type client struct {
	exec.Destroyable
	dirPath string
}

func newClient(destroyCallback func() error, dirPath string) *client {
	return &client{exec.NewDestroyable(destroyCallback), dirPath}
}

func (this *client) DirPath() string {
	return this.dirPath
}

func (this *client) Execute(cmd *exec.Cmd) func() error {
	value, err := this.Do(func() (interface{}, error) {
		stdosexecCmd, err := this.stdosexecCmd(cmd)
		if err != nil {
			return nil, err
		}
		if err := stdosexecCmd.Start(); err != nil {
			return nil, err
		}
		return func() error { return stdosexecCmd.Wait() }, nil
	})
	if err != nil {
		return func() error { return err }
	}
	return value.(func() error)
}

func (this *client) IsFileExists(path string) (bool, error) {
	if err := this.validatePath(path); err != nil {
		return false, err
	}
	value, err := this.Do(func() (interface{}, error) {
		_, err := stdos.Stat(this.absolutePath(path))
		if err == nil {
			return true, nil
		}
		if stdos.IsNotExist(err) {
			return false, nil
		}
		return false, err
	})
	if err != nil {
		return false, err
	}
	return value.(bool), nil
}

func (this *client) Open(path string) (exec.ReadFile, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		exists, err := this.IsFileExists(path)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, exec.ErrFileDoesNotExist
		}
		return stdos.Open(this.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*stdos.File), nil
}

func (this *client) Create(path string) (exec.WriteFile, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		return stdos.Create(this.absolutePath(path))
	})
	if err != nil {
		return nil, err
	}
	return value.(*stdos.File), nil
}

func (this *client) MkdirAll(path string, perm stdos.FileMode) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	_, err := this.Do(func() (interface{}, error) {
		return nil, stdos.MkdirAll(this.absolutePath(path), perm)
	})
	return err
}

func (this *client) ListRegularFiles(path string) ([]string, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		files := make([]string, 0)
		err := filepath.Walk(
			this.absolutePath(path),
			func(path string, info stdos.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.Mode().IsRegular() {
					relativeFile, err := filepath.Rel(this.dirPath, path)
					if err != nil {
						return err
					}
					files = append(files, relativeFile)
				}
				return nil
			},
		)
		if err != nil {
			return nil, err
		}
		return files, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]string), nil
}

func (this *client) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (this *client) Match(pattern string, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

func (this *client) ToSlash(path string) string {
	return filepath.ToSlash(path)
}

func (this *client) PathSeparator() string {
	return string(stdos.PathSeparator)
}

func (this *client) NewSubDirExecutorReadFileManager(path string) (exec.ExecutorReadFileManager, error) {
	return this.newSubDirClient(path)
}

func (this *client) NewSubDirExecutorWriteFileManager(path string) (exec.ExecutorWriteFileManager, error) {
	return this.newSubDirClient(path)
}

func (this *client) NewSubDirClient(path string) (exec.Client, error) {
	return this.newSubDirClient(path)
}

func (this *client) newSubDirClient(path string) (*client, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	exists, err := this.IsFileExists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exec.ErrFileAlreadyExists
	}
	if err := stdos.Mkdir(this.absolutePath(path), 0755); err != nil {
		return nil, err
	}
	subDirClient := newClient(func() error { return this.removeDir(path) }, this.absolutePath(path))
	if err := this.AddChild(subDirClient); err != nil {
		return nil, err
	}
	return subDirClient, nil
}

func (this *client) removeDir(path string) error {
	if err := this.validateIsDir(path); err != nil {
		return err
	}
	return stdos.RemoveAll(this.absolutePath(path))
}

func (this *client) validateIsDir(path string) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	fileInfo, err := stdos.Stat(this.absolutePath(path))
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return exec.ErrFileDoesNotExist
	}
	return nil
}

func (this *client) validatePath(path string) error {
	if filepath.IsAbs(path) {
		return exec.ErrNotRelativePath
	}
	// TODO(pedge): EvalSymlinks fails if the file does not exist
	//path, err := filepath.EvalSymlinks(filepath.Clean(this.absolutePath(path)))
	//if err != nil {
	//return err
	//}
	//if !strings.HasPrefix(path, this.DirPath()) {
	//return exec.ErrPathOutOfContext
	//}
	return nil
}

func (this *client) absolutePath(path string) string {
	return this.Join(this.dirPath, path)
}

func (this *client) stdosexecCmd(cmd *exec.Cmd) (*stdosexec.Cmd, error) {
	if cmd.Args == nil || len(cmd.Args) == 0 {
		return nil, exec.ErrArgsEmpty
	}
	var stdosexecCmd *stdosexec.Cmd
	if len(cmd.Args) == 1 {
		stdosexecCmd = stdosexec.Command(cmd.Args[0])
	} else {
		stdosexecCmd = stdosexec.Command(cmd.Args[0], cmd.Args[1:]...)
	}
	stdosexecCmd.Dir = this.dirPath
	stdosexecCmd.Stdin = cmd.Stdin
	stdosexecCmd.Stdout = cmd.Stdout
	stdosexecCmd.Stderr = cmd.Stderr
	return stdosexecCmd, nil
}
