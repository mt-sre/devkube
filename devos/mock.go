//nolint:revive
package devos

import (
	"context"
	"io"
	"io/fs"

	"github.com/stretchr/testify/mock"
	"k8s.io/utils/exec"
)

type MockFS struct {
	mock.Mock
}

func (m *MockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	args := m.Called(name)
	return args.Get(0).([]fs.DirEntry), args.Error(1)
}

func (m *MockFS) MkdirAll(path string, perm fs.FileMode) error {
	return m.Called(path, perm).Error(0)
}

func (m *MockFS) ReadFile(name string) ([]byte, error) {
	args := m.Called(name)
	return args.Get(0).([]byte), args.Error(1)
}

type MockExec struct {
	mock.Mock
}

func (e *MockExec) Command(cmd string, args ...string) exec.Cmd {
	return e.Called(cmd, args).Get(0).(exec.Cmd)
}

func (e *MockExec) CommandContext(ctx context.Context, cmd string, args ...string) exec.Cmd {
	return e.Called(ctx, cmd, args).Get(0).(exec.Cmd)
}

func (e *MockExec) LookPath(file string) (string, error) {
	args := e.Called(file)
	return args.String(0), args.Error(1)
}

type MockCmd struct {
	mock.Mock
}

func (c *MockCmd) Run() error { return c.Called().Error(0) }
func (c *MockCmd) CombinedOutput() ([]byte, error) {
	args := c.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (c *MockCmd) Output() ([]byte, error) {
	args := c.Called()
	return args.Get(0).([]byte), args.Error(1)
}
func (c *MockCmd) SetDir(dir string)       { c.Called(dir) }
func (c *MockCmd) SetStdin(in io.Reader)   { c.Called(in) }
func (c *MockCmd) SetStdout(out io.Writer) { c.Called(out) }
func (c *MockCmd) SetStderr(out io.Writer) { c.Called(out) }
func (c *MockCmd) SetEnv(env []string)     { c.Called(env) }
func (c *MockCmd) StdoutPipe() (io.ReadCloser, error) {
	args := c.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (c *MockCmd) StderrPipe() (io.ReadCloser, error) {
	args := c.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (c *MockCmd) Start() error { return c.Called().Error(0) }
func (c *MockCmd) Wait() error  { return c.Called().Error(0) }
func (c *MockCmd) Stop()        { c.Called() }
