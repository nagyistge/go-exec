package os

import (
	"bytes"
	stdos "os"
	"strings"

	"testing"

	"github.com/peter-edge/exec"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	clientProvider exec.ClientProvider
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (this *Suite) SetupSuite() {
}

func (this *Suite) SetupTest() {
	this.clientProvider = newClientProvider()
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
			if err == exec.ErrAlreadyDestroyed {
				count++
			} else {
				require.NoError(this.T(), err)
			}
		}
	}
	require.Equal(this.T(), 9, count)
}

func (this *Suite) newClient() exec.Client {
	client, err := this.clientProvider.NewTempDirClient()
	require.NoError(this.T(), err)
	this.checkFileExists(client.DirPath())
	return client
}

func (this *Suite) execute(client exec.Client, args []string) (stdout string, stderr string) {
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	err := client.Execute(
		&exec.Cmd{
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

func (this *Suite) destroy(client exec.Client) {
	err := client.Destroy()
	require.NoError(this.T(), err)
	this.checkFileDoesNotExist(client.DirPath())
}

func (this *Suite) checkFileExists(path string) {
	_, err := stdos.Stat(path)
	require.NoError(this.T(), err)
}

func (this *Suite) checkFileDoesNotExist(path string) {
	_, err := stdos.Stat(path)
	require.True(this.T(), stdos.IsNotExist(err))
}
