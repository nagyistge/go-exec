package exec

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"testing"

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

func (this *Suite) SetupSuite() {
}

func (this *Suite) SetupTest() {
	this.clientProvider = newOsClientProvider()
}

func (this *Suite) TearDownTest() {
	require.NoError(this.T(), this.clientProvider.Destroy())
}

func (this *Suite) TearDownSuite() {
}

func (this *Suite) TestPwd() {
	client := this.newClient()
	pwd, _ := this.execute(client, []string{"pwd", "-P"})
	require.Equal(this.T(), client.DirPath(), pwd)
	this.destroy(client)
}

func (this *Suite) TestEnv() {
	client := this.newClient()
	writeFile, err := client.Create("echo_foo.sh")
	require.NoError(this.T(), err)
	fromFile, err := os.Open("_testdata/echo_foo.sh")
	require.NoError(this.T(), err)
	defer fromFile.Close()
	data, err := ioutil.ReadAll(fromFile)
	require.NoError(this.T(), err)
	_, err = writeFile.Write(data)
	require.NoError(this.T(), err)
	err = writeFile.Chmod(0777)
	require.NoError(this.T(), err)
	writeFile.Close()

	var output bytes.Buffer
	err = client.Execute(
		&Cmd{
			Args:   []string{"bash", "echo_foo.sh"},
			Env:    []string{"FOO=foo"},
			Stdout: &output,
		},
	)()
	require.NoError(this.T(), err)
	require.Equal(this.T(), "foo", strings.TrimSpace(output.String()))
}

func (this *Suite) TestPipe() {
	client := this.newClient()
	var input bytes.Buffer
	input.WriteString("hello\n")
	input.WriteString("hello\n")
	input.WriteString("woot\n")
	input.WriteString("hello\n")
	input.WriteString("foo\n")
	input.WriteString("woot\n")
	input.WriteString("foo\n")
	input.WriteString("woot\n")
	input.WriteString("hello\n")
	input.WriteString("foo\n")
	input.WriteString("foo\n")
	input.WriteString("foo\n")
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
	require.NoError(this.T(), err)
	require.True(this.T(), strings.Contains(output.String(), "3"))
	this.destroy(client)
}

func (this *Suite) TestListFileInfosShallow() {
	client := this.newClient()
	err := client.MkdirAll("dirOne", 0755)
	require.NoError(this.T(), err)
	err = client.MkdirAll("dirTwo", 0755)
	require.NoError(this.T(), err)
	err = client.MkdirAll("dirOne/dirOneOne", 0755)
	require.NoError(this.T(), err)
	err = client.MkdirAll("dirTwo/dirTwoOne", 0755)
	require.NoError(this.T(), err)
	file, err := client.Create("one")
	require.NoError(this.T(), err)
	file.Close()
	file, err = client.Create("two")
	require.NoError(this.T(), err)
	file.Close()
	file, err = client.Create("dirOne/oneOne")
	require.NoError(this.T(), err)
	file.Close()

	fileNameToDir := map[string]bool{
		"dirOne": true,
		"dirTwo": true,
		"one":    false,
		"two":    false,
	}
	fileInfos, err := client.ListFileInfosShallow(".")
	require.NoError(this.T(), err)
	require.Equal(this.T(), 4, len(fileInfos))
	for _, fileInfo := range fileInfos {
		dir, ok := fileNameToDir[fileInfo.Name()]
		require.True(this.T(), ok)
		require.Equal(this.T(), dir, fileInfo.IsDir())
		require.Equal(this.T(), !dir, fileInfo.Mode().IsRegular())
	}
}

func (this *Suite) TestLotsOfDestroys() {
	client := this.newClient()
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
			if err == ErrAlreadyDestroyed {
				count++
			} else {
				require.NoError(this.T(), err)
			}
		}
	}
	require.Equal(this.T(), 9, count)
}

func (this *Suite) newClient() Client {
	client, err := this.clientProvider.NewTempDirClient()
	require.NoError(this.T(), err)
	this.checkFileExists(client.DirPath())
	return client
}

func (this *Suite) execute(client Client, args []string) (stdout string, stderr string) {
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
	require.NoError(this.T(), err, stderr)
	return
}

func (this *Suite) destroy(client Client) {
	err := client.Destroy()
	require.NoError(this.T(), err)
	this.checkFileDoesNotExist(client.DirPath())
}

func (this *Suite) checkFileExists(path string) {
	_, err := os.Stat(path)
	require.NoError(this.T(), err)
}

func (this *Suite) checkFileDoesNotExist(path string) {
	_, err := os.Stat(path)
	require.True(this.T(), os.IsNotExist(err))
}
