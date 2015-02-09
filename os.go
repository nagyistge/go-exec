package exec

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	stdosexec "os/exec"
	"path/filepath"
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

const (
	tempDirPrefix         = "exec"
	readDirNamesSliceSize = 100
)

type osClientProvider struct {
	Destroyable
	execOptions *OsExecOptions
}

func newOsExecutorReadFileManagerProvider(execOptions *OsExecOptions) *osClientProvider {
	return newOsClientProvider(execOptions)
}

func newOsExecutorWriteFileManagerProvider(execOptions *OsExecOptions) *osClientProvider {
	return newOsClientProvider(execOptions)
}

func newOsClientProvider(execOptions *OsExecOptions) *osClientProvider {
	return &osClientProvider{NewDestroyable(nil), execOptions}
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
	baseDirPath, err := this.baseDirPath()
	if err != nil {
		return nil, err
	}
	client := newOsClient(func() error { return this.removeTempDir(tempDir) }, tempDir, baseDirPath)
	if err := this.AddChild(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (this *osClientProvider) baseDirPath() (string, error) {
	if this.execOptions.TmpDir == "" {
		return "", nil
	}
	return filepath.EvalSymlinks(filepath.Clean(this.execOptions.TmpDir))
}

func (this *osClientProvider) createTempDir() (string, error) {
	value, err := this.Do(func() (interface{}, error) {
		var tempDir string
		var err error
		baseDirPath, err := this.baseDirPath()
		if err != nil {
			return nil, err
		}
		if baseDirPath == "" {
			tempDir, err = ioutil.TempDir("", tempDirPrefix)
			if err != nil {
				return "", err
			}
		} else {
			tempDir = filepath.Join(baseDirPath, uuid.NewUUID().String())
			if err := os.Mkdir(tempDir, 0755); err != nil {
				return "", err
			}
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
func (this *osClientProvider) removeTempDir(tempDir string) error {
	if err := this.validateIsDir(tempDir); err != nil {
		return err
	}
	return os.RemoveAll(tempDir)
}

// this is only called in thread-safe context
func (this *osClientProvider) validateIsDir(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return ErrFileDoesNotExist
	}
	return nil
}

type osClient struct {
	Destroyable
	dirPath     string
	baseDirPath string
}

func newOsClient(destroyCallback func() error, dirPath string, baseDirPath string) *osClient {
	return &osClient{NewDestroyable(destroyCallback), dirPath, baseDirPath}
}

func (this *osClient) DirPath() string {
	return this.dirPath
}

func (this *osClient) BaseDirPath() (string, bool) {
	if this.baseDirPath == "" {
		return "", false
	}
	return this.baseDirPath, true
}

func (this *osClient) InsideContainer() (bool, error) {
	value, err := this.Do(func() (interface{}, error) {
		containerId, err := this.containerId()
		if err != nil {
			return nil, err
		}
		return containerId != "", nil
	})
	if err != nil {
		return false, err
	}
	return value.(bool), nil
}

func (this *osClient) ContainerId() (string, error) {
	value, err := this.Do(func() (interface{}, error) {
		containerId, err := this.containerId()
		if err != nil {
			return "", err
		}
		if containerId == "" {
			return "", errors.New("exec: not inside a ontainer")
		}
		return containerId, nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (this *osClient) containerId() (string, error) {
	lines, err := this.cgroupDockerLines()
	if err != nil {
		return "", err
	}
	if lines == nil || len(lines) == 0 {
		return "", nil
	}
	// TODO(pedge)
	containerId := strings.TrimPrefix(strings.TrimPrefix(lines[0], "/docker/"), "/system.slice/docker-")
	for i := 1; i < len(lines); i++ {
		anotherContainerId := strings.TrimPrefix(strings.TrimPrefix(lines[i], "/docker/"), "/system.slice/docker-")
		if containerId != anotherContainerId {
			return "", fmt.Errorf("exec: recieved mismatched container Ids: %v %v", containerId, anotherContainerId)
		}
	}
	return containerId, nil
}

func (this *osClient) cgroupDockerLines() ([]string, error) {
	lines, err := readFileLines("/proc/self/cgroup")
	if err != nil {
		return nil, err
	}
	dockerLines := make([]string, 0)
	for _, line := range lines {
		split := strings.Split(line, ":")
		if len(split) > 2 {
			if strings.HasPrefix(split[2], "/docker/") || strings.HasPrefix(split[2], "/system.slice/docker-") {
				dockerLines = append(dockerLines, split[2])
			}
		}
	}
	return dockerLines, nil
}

func (this *osClient) Execute(cmd *Cmd) func() error {
	if cmd.SubDir != "" {
		if err := this.validatePath(cmd.SubDir); err != nil {
			return func() error { return err }
		}
	}
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

func (this *osClient) ExecutePiped(pipeCmdList *PipeCmdList) func() error {
	numCmds := len(pipeCmdList.PipeCmds)
	if numCmds < 2 {
		return func() error { return ErrNotMultipleCommands }
	}
	for _, pipeCmd := range pipeCmdList.PipeCmds {
		if pipeCmd.SubDir != "" {
			if err := this.validatePath(pipeCmd.SubDir); err != nil {
				return func() error { return err }
			}
		}
	}
	stdosexecCmds := make([]*stdosexec.Cmd, numCmds)
	for i, pipeCmd := range pipeCmdList.PipeCmds {
		stdosexecCmd, err := this.stdosexecPipeCmd(pipeCmd)
		if err != nil {
			return func() error { return err }
		}
		stdosexecCmds[i] = stdosexecCmd
	}
	readers := make([]*io.PipeReader, numCmds-1)
	writers := make([]*io.PipeWriter, numCmds-1)
	value, err := this.Do(func() (interface{}, error) {
		reader, writer := io.Pipe()
		readers[0] = reader
		writers[0] = writer
		stdosexecCmds[0].Stdin = pipeCmdList.Stdin
		for i := 0; i < numCmds-1; i++ {
			stdosexecCmds[i].Stdout = writer
			stdosexecCmds[i].Stderr = pipeCmdList.Stderr
			stdosexecCmds[i+1].Stdin = reader
			if i != numCmds-2 {
				reader, writer = io.Pipe()
				readers[i+1] = reader
				writers[i+1] = writer
			}
		}
		stdosexecCmds[numCmds-1].Stdout = pipeCmdList.Stdout
		stdosexecCmds[numCmds-1].Stderr = pipeCmdList.Stderr
		for _, stdosexecCmd := range stdosexecCmds {
			if err := stdosexecCmd.Start(); err != nil {
				return nil, err
			}
		}
		return func() error {
			for i := 0; i < numCmds-1; i++ {
				if err := stdosexecCmds[i].Wait(); err != nil {
					return err
				}
				if i != 0 {
					if err := readers[i-1].Close(); err != nil {
						return err
					}
				}
				if err := writers[i].Close(); err != nil {
					return err
				}
			}
			if err := stdosexecCmds[numCmds-1].Wait(); err != nil {
				return err
			}
			if err := readers[numCmds-2].Close(); err != nil {
				return err
			}
			return nil
		}, nil
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
		return isFileExists(this.absolutePath(path))
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
		exists, err := this.IsFileExists(path)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrFileDoesNotExist
		}
		return os.Open(this.absolutePath(path))
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
		return os.Create(this.absolutePath(path))
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
		return nil, os.MkdirAll(this.absolutePath(path), perm)
	})
	return err
}

func (this *osClient) ListRegularFiles(path string) ([]string, error) {
	if err := this.validatePath(path); err != nil {
		return nil, err
	}
	value, err := this.Do(func() (interface{}, error) {
		files := make([]string, 0)
		err := filepath.Walk(
			this.absolutePath(path),
			func(path string, info os.FileInfo, err error) error {
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

func (this *osClient) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (this *osClient) Match(pattern string, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

func (this *osClient) ToSlash(path string) string {
	return filepath.ToSlash(path)
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
	if err := os.Mkdir(this.absolutePath(path), 0755); err != nil {
		return nil, err
	}
	subDirClient := newOsClient(func() error { return this.removeDir(path) }, this.absolutePath(path), this.baseDirPath)
	if err := this.AddChild(subDirClient); err != nil {
		return nil, err
	}
	return subDirClient, nil
}

func (this *osClient) removeDir(path string) error {
	if err := this.validateIsDir(path); err != nil {
		return err
	}
	return os.RemoveAll(this.absolutePath(path))
}

func (this *osClient) validateIsDir(path string) error {
	if err := this.validatePath(path); err != nil {
		return err
	}
	fileInfo, err := os.Stat(this.absolutePath(path))
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
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

func (this *osClient) stdosexecCmd(cmd *Cmd) (*stdosexec.Cmd, error) {
	if cmd.Args == nil || len(cmd.Args) == 0 {
		return nil, ErrArgsEmpty
	}
	var stdosexecCmd *stdosexec.Cmd
	if len(cmd.Args) == 1 {
		stdosexecCmd = stdosexec.Command(cmd.Args[0])
	} else {
		stdosexecCmd = stdosexec.Command(cmd.Args[0], cmd.Args[1:]...)
	}
	if cmd.SubDir != "" {
		stdosexecCmd.Dir = this.absolutePath(cmd.SubDir)
	} else {
		stdosexecCmd.Dir = this.dirPath
	}
	if cmd.Env != nil && len(cmd.Env) > 0 {
		stdosexecCmd.Env = cmd.Env
	}
	stdosexecCmd.Stdin = cmd.Stdin
	stdosexecCmd.Stdout = cmd.Stdout
	stdosexecCmd.Stderr = cmd.Stderr
	return stdosexecCmd, nil
}

func (this *osClient) stdosexecPipeCmd(pipeCmd *PipeCmd) (*stdosexec.Cmd, error) {
	if pipeCmd.Args == nil || len(pipeCmd.Args) == 0 {
		return nil, ErrArgsEmpty
	}
	var stdosexecCmd *stdosexec.Cmd
	if len(pipeCmd.Args) == 1 {
		stdosexecCmd = stdosexec.Command(pipeCmd.Args[0])
	} else {
		stdosexecCmd = stdosexec.Command(pipeCmd.Args[0], pipeCmd.Args[1:]...)
	}
	if pipeCmd.SubDir != "" {
		stdosexecCmd.Dir = this.absolutePath(pipeCmd.SubDir)
	} else {
		stdosexecCmd.Dir = this.dirPath
	}
	if pipeCmd.Env != nil && len(pipeCmd.Env) > 0 {
		stdosexecCmd.Env = pipeCmd.Env
	}
	return stdosexecCmd, nil
}

func readFileLines(path string) ([]string, error) {
	reader, err := readFile(path)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	if reader == nil {
		return lines, nil
	}
	bufReader := bufio.NewReader(reader)
	for line, err := bufReader.ReadString('\n'); err != io.EOF; line, err = bufReader.ReadString('\n') {
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func readFile(path string) (retReader io.Reader, retErr error) {
	exists, err := isFileExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, file); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func isFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
