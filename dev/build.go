package dev

import (
	"os"
	"os/exec"
)

type (
	CommandImage struct {
		Push       bool
		BinaryName string
	}
	PackageImage struct {
		ExtraDeps  []interface{}
		Push       bool
		SourcePath string
	}
	ImageBuildInfo struct {
		ImageName    string
		ImageTag     string
		CmdImageOpts *CommandImage
		PkgImageOpts *PackageImage
		CacheDir     string
	}
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// BuildCmdImage is a generic image build function, when the image just relies
// on a static binary build from cmd/*,
// requires the binaries to be built beforehand
func BuildCmdImage(buildInfo ImageBuildInfo, populateCache func() error) {
	must(populateCache())

	containerRuntime, err := DetectContainerRuntime()
	must(err)

	cmds := [][]string{
		{string(containerRuntime), "build", "-t", buildInfo.ImageTag, "-f", "Containerfile", "."},
		{string(containerRuntime), "image", "save", "-o", buildInfo.CacheDir + ".tar", buildInfo.ImageTag},
	}

	// Build image!
	for _, command := range cmds {
		buildCmd := exec.Command(command[0], command[1:]...)
		buildCmd.Stderr = os.Stderr
		buildCmd.Stdout = os.Stdout
		buildCmd.Dir = buildInfo.CacheDir
		must(buildCmd.Run())
	}
}
