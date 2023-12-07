//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const golangciLintVersion = "v1.55.0"

var (
	cacheDir = filepath.Join(".", ".cache")
	binDir   = filepath.Join(cacheDir, "bin")
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	absBinDir, err := filepath.Abs(binDir)
	must(err)
	os.Setenv("PATH", absBinDir+":"+os.Getenv("PATH"))
	os.Setenv("GOBIN", absBinDir)
}

func Unit() {
	must(sh.RunWithV(
		map[string]string{"CGO_ENABLED": "1"},
		"go", "test", "-v", "-race",
		fmt.Sprintf("-coverprofile=%s", filepath.Join(cacheDir, "cov.out")),
		"./...",
	))
}

func Lint() {
	must(sh.RunV(mg.GoCmd(), "install", fmt.Sprintf("github.com/golangci/golangci-lint/cmd/golangci-lint@%s", golangciLintVersion)))
	must(sh.RunV("golangci-lint", "run", "./...", "--deadline=15m"))
}
