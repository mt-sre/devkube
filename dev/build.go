package dev

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"os/exec"
	"path/filepath"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// generic image build function, when the image just relies on
// a static binary build from cmd/*
func buildCmdImage(imageName string) {
	opts, ok := commandImages[imageName]
	if !ok {
		panic(fmt.Sprintf("unknown cmd image: %s", imageName))
	}
	cmd := imageName
	if len(opts.BinaryName) != 0 {
		cmd = opts.BinaryName
	}

	mg.Deps(mg.F(Build.Binary, cmd, linuxAMD64Arch.OS, linuxAMD64Arch.Arch), mg.F(Build.cleanImageCacheDir, imageName))

	imageCacheDir := locations.ImageCache(imageName)
	imageTag := locations.ImageURL(imageName, false)

	// prepare build context
	must(sh.Copy(filepath.Join(imageCacheDir, cmd), locations.binaryDst(cmd, linuxAMD64Arch)))
	must(sh.Copy(filepath.Join(imageCacheDir, "Containerfile"), filepath.Join("config", "images", imageName+".Containerfile")))
	must(sh.Copy(filepath.Join(imageCacheDir, "passwd"), filepath.Join("config", "images", "passwd")))

	containerRuntime := locations.ContainerRuntime()
	cmds := [][]string{
		{containerRuntime, "build", "-t", imageTag, "-f", "Containerfile", "."},
		{containerRuntime, "image", "save", "-o", imageCacheDir + ".tar", imageTag},
	}

	// Build image!
	for _, command := range cmds {
		buildCmd := exec.Command(command[0], command[1:]...)
		buildCmd.Stderr = os.Stderr
		buildCmd.Stdout = os.Stdout
		buildCmd.Dir = imageCacheDir
		must(buildCmd.Run())
	}
}
