package devos

import (
	"io/fs"
	"os"

	"github.com/stretchr/testify/mock"
)

type FS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	ReadFile(name string) ([]byte, error)
}

func RealFSIfUnset(fs FS) FS {
	if fs == nil {
		return RealFS{}
	}
	return fs
}

type RealFS struct{}

func (RealFS) ReadDir(name string) ([]fs.DirEntry, error)   { return os.ReadDir(name) }
func (RealFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (RealFS) ReadFile(name string) ([]byte, error)         { return os.ReadFile(name) }

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
