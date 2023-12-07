package devos

import (
	"io/fs"
	"os"
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
