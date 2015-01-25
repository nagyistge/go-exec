package os

import (
	"io/ioutil"
	stdos "os"
	"path/filepath"

	"gopkg.in/peter-edge/exec.v1"
)

const (
	tempDirPrefix = "exec"
)

func NewExecutorReadFileManagerProvider() (exec.ExecutorReadFileManagerProvider, error) {
	return newClientProvider(), nil
}

func NewExecutorWriteFileManagerProvider() (exec.ExecutorWriteFileManagerProvider, error) {
	return newClientProvider(), nil
}

func NewClientProvider() (exec.ClientProvider, error) {
	return newClientProvider(), nil
}

type clientProvider struct {
	exec.Destroyable
}

func newClientProvider() *clientProvider {
	return &clientProvider{exec.NewDestroyable(nil)}
}

func (this *clientProvider) NewTempDirExecutorReadFileManager() (exec.ExecutorReadFileManager, error) {
	return this.NewTempDirClient()
}

func (this *clientProvider) NewTempDirExecutorWriteFileManager() (exec.ExecutorWriteFileManager, error) {
	return this.NewTempDirClient()
}

func (this *clientProvider) NewTempDirClient() (exec.Client, error) {
	tempDir, err := this.createTempDir()
	if err != nil {
		return nil, err
	}
	client := newClient(func() error { return this.removeTempDir(tempDir) }, tempDir)
	if err := this.AddChild(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (this *clientProvider) createTempDir() (string, error) {
	value, err := this.Do(func() (interface{}, error) {
		tempDir, err := ioutil.TempDir("", tempDirPrefix)
		if err != nil {
			return "", err
		}
		tempDir, err = filepath.EvalSymlinks(filepath.Clean(tempDir))
		if err != nil {
			return "", err
		}
		return tempDir, nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

// this is only called in thread-safe context
func (this *clientProvider) removeTempDir(tempDir string) error {
	if err := this.validateIsDir(tempDir); err != nil {
		return err
	}
	return stdos.RemoveAll(tempDir)
}

// this is only called in thread-safe context
func (this *clientProvider) validateIsDir(path string) error {
	fileInfo, err := stdos.Stat(path)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return exec.ErrFileDoesNotExist
	}
	return nil
}
