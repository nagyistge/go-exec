package exec

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"testing"

	"github.com/codeship/go-concurrent"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	clientProvider ClientProvider
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
}

func (s *Suite) SetupTest() {
	s.clientProvider = newOsClientProvider(&OsExecOptions{})
}

func (s *Suite) TearDownTest() {
	require.NoError(s.T(), s.clientProvider.Destroy())
}

func (s *Suite) TearDownSuite() {
}

func (s *Suite) TestPwd() {
	client := s.newClient()
	pwd, _ := s.execute(client, []string{"pwd", "-P"})
	require.Equal(s.T(), client.DirPath(), pwd)
	s.destroy(client)
}

func (s *Suite) TestEnv() {
	client := s.newClient()
	writeFile, err := client.Create("echo_foo.sh")
	require.NoError(s.T(), err)
	fromFile, err := os.Open("_testdata/echo_foo.sh")
	require.NoError(s.T(), err)
	defer s.checkClose(fromFile)
	data, err := ioutil.ReadAll(fromFile)
	require.NoError(s.T(), err)
	_, err = writeFile.Write(data)
	require.NoError(s.T(), err)
	err = writeFile.Chmod(0777)
	require.NoError(s.T(), err)
	s.checkClose(writeFile)

	var output bytes.Buffer
	err = client.Execute(
		&Cmd{
			Args:   []string{"bash", "echo_foo.sh"},
			Env:    []string{"FOO=foo"},
			Stdout: &output,
		},
	)()
	require.NoError(s.T(), err)
	require.Equal(s.T(), "foo", strings.TrimSpace(output.String()))
}

func (s *Suite) TestPipe() {
	client := s.newClient()
	var input bytes.Buffer
	_, _ = input.WriteString("hello\n")
	_, _ = input.WriteString("hello\n")
	_, _ = input.WriteString("woot\n")
	_, _ = input.WriteString("hello\n")
	_, _ = input.WriteString("foo\n")
	_, _ = input.WriteString("woot\n")
	_, _ = input.WriteString("foo\n")
	_, _ = input.WriteString("woot\n")
	_, _ = input.WriteString("hello\n")
	_, _ = input.WriteString("foo\n")
	_, _ = input.WriteString("foo\n")
	_, _ = input.WriteString("foo\n")
	var output bytes.Buffer
	err := client.ExecutePiped(
		&PipeCmdList{
			PipeCmds: []*PipeCmd{
				&PipeCmd{
					Args: []string{"sort"},
				},
				&PipeCmd{
					Args: []string{"uniq"},
				},
				&PipeCmd{
					Args: []string{"wc", "-l"},
				},
			},
			Stdin:  &input,
			Stdout: &output,
		},
	)()
	require.NoError(s.T(), err)
	require.True(s.T(), strings.Contains(output.String(), "3"))
	s.destroy(client)
}

func (s *Suite) TestListFileInfosShallow() {
	client := s.newClient()
	err := client.MkdirAll("dirOne", 0755)
	require.NoError(s.T(), err)
	err = client.MkdirAll("dirTwo", 0755)
	require.NoError(s.T(), err)
	err = client.MkdirAll("dirOne/dirOneOne", 0755)
	require.NoError(s.T(), err)
	err = client.MkdirAll("dirTwo/dirTwoOne", 0755)
	require.NoError(s.T(), err)
	file, err := client.Create("one")
	require.NoError(s.T(), err)
	s.checkClose(file)
	file, err = client.Create("two")
	require.NoError(s.T(), err)
	s.checkClose(file)
	file, err = client.Create("dirOne/oneOne")
	require.NoError(s.T(), err)
	s.checkClose(file)

	fileNameToDir := map[string]bool{
		"dirOne": true,
		"dirTwo": true,
		"one":    false,
		"two":    false,
	}
	dir, err := client.Open(".")
	require.NoError(s.T(), err)
	fileInfos, err := dir.Readdir(-1)
	require.NoError(s.T(), err)
	s.checkClose(dir)
	require.Equal(s.T(), 4, len(fileInfos))
	for _, fileInfo := range fileInfos {
		dir, ok := fileNameToDir[fileInfo.Name()]
		require.True(s.T(), ok)
		require.Equal(s.T(), dir, fileInfo.IsDir())
		require.Equal(s.T(), !dir, fileInfo.Mode().IsRegular())
	}
}

func (s *Suite) TestLotsOfDestroys() {
	client := s.newClient()
	done := make(chan error)
	for i := 0; i < 10; i++ {
		go func() {
			done <- client.Destroy()
		}()
	}
	count := 0
	for i := 0; i < 10; i++ {
		err := <-done
		if err != nil {
			if err == concurrent.ErrAlreadyDestroyed {
				count++
			} else {
				require.NoError(s.T(), err)
			}
		}
	}
	require.Equal(s.T(), 9, count)
}

func (s *Suite) newClient() Client {
	client, err := s.clientProvider.NewTempDirClient()
	require.NoError(s.T(), err)
	s.checkFileExists(client.DirPath())
	return client
}

func (s *Suite) execute(client Client, args []string) (stdout string, stderr string) {
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	err := client.Execute(
		&Cmd{
			Args:   args,
			Stdout: &stdoutBuffer,
			Stderr: &stderrBuffer,
		},
	)()
	stdout = strings.TrimSpace(stdoutBuffer.String())
	stderr = strings.TrimSpace(stderrBuffer.String())
	require.NoError(s.T(), err, stderr)
	return
}

func (s *Suite) destroy(client Client) {
	err := client.Destroy()
	require.NoError(s.T(), err)
	s.checkFileDoesNotExist(client.DirPath())
}

func (s *Suite) checkFileExists(path string) {
	_, err := os.Stat(path)
	require.NoError(s.T(), err)
}

func (s *Suite) checkFileDoesNotExist(path string) {
	_, err := os.Stat(path)
	require.True(s.T(), os.IsNotExist(err))
}

func (s *Suite) checkClose(closer io.Closer) {
	err := closer.Close()
	require.NoError(s.T(), err)
}
