package impl

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

	clientProvider *clientProvider
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
}

func (this *Suite) TearDownSuite() {
}

func (this *Suite) TestSimple() {
	client, err := this.clientProvider.NewTempDirClient()
	require.NoError(this.T(), err)
	_, err = stdos.Stat(client.DirPath())
	require.NoError(this.T(), err)
	var buffer bytes.Buffer
	err = client.Execute(
		&exec.Cmd{
			Args:   []string{"pwd", "-P"},
			Stdout: &buffer,
		},
	)()
	require.NoError(this.T(), err)
	dirString := strings.TrimSpace(buffer.String())
	require.Equal(this.T(), client.DirPath(), dirString)
	err = this.clientProvider.Destroy()
	require.NoError(this.T(), err)
	_, err = stdos.Stat(client.DirPath())
	require.True(this.T(), stdos.IsNotExist(err), "Directory still exists")
}
