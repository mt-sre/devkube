package dev

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ImageBuildInfo struct {
	ImageTag      string
	CacheDir      string
	ContainerFile string
	ContextDir    string
	Runtime       string
}

type PackageBuildInfo struct {
	ImageTag   string
	CacheDir   string
	SourcePath string // source directory
	OutputPath string // destination: .tar file path
	Runtime    string
}

type ImagePushInfo struct {
	ImageTag   string
	CacheDir   string
	Runtime    string
	DigestFile string
}

// execCommand is replaced with helper function when testing
var execCommand = exec.Command

func execError(command []string, err error) error {
	return fmt.Errorf("running command '%s': %w", strings.Join(command, " "), err)
}

func newExecCmd(args []string, cacheDir string) *exec.Cmd {
	cmd := execCommand(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = cacheDir
	return cmd
}

// BuildImage is a generic image build function,
// requires the binaries to be built beforehand
func BuildImage(buildInfo *ImageBuildInfo, deps []interface{}) error {
	if len(deps) > 0 {
		mg.SerialDeps(deps...)
	}

	buildCmdArgs := []string{buildInfo.Runtime, "build", "-t", buildInfo.ImageTag}
	if buildInfo.ContainerFile != "" {
		buildCmdArgs = append(buildCmdArgs, "-f", buildInfo.ContainerFile)
	}
	buildCmdArgs = append(buildCmdArgs, buildInfo.ContextDir)

	commands := [][]string{
		buildCmdArgs,
		{buildInfo.Runtime, "image", "save", "-o", buildInfo.CacheDir + ".tar", buildInfo.ImageTag},
	}

	// Build image!
	for _, command := range commands {
		buildCmd := newExecCmd(command, buildInfo.CacheDir)
		if err := buildCmd.Run(); err != nil {
			return execError(command, err)
		}
	}
	return nil
}

// BuildPackage builds a package image using the package operator CLI,
// requires `kubectl package` command to be available on the system
func BuildPackage(buildInfo *PackageBuildInfo, deps []interface{}) error {
	if len(deps) > 0 {
		mg.SerialDeps(deps...)
	}

	buildArgs := []string{
		"kubectl", "package", "build", "--tag", buildInfo.ImageTag,
		"--output", buildInfo.OutputPath, buildInfo.SourcePath,
	}
	importArgs := []string{
		buildInfo.Runtime, "import", buildInfo.OutputPath, buildInfo.ImageTag,
	}

	for _, args := range [][]string{buildArgs, importArgs} {
		command := newExecCmd(args, buildInfo.CacheDir)
		if err := command.Run(); err != nil {
			return execError(args, err)
		}
	}
	return nil
}

func quayLogin(runtime, cacheDir string) error {
	args := []string{runtime, "login", "-u=" + os.Getenv("QUAY_USER"), "-p=" + os.Getenv("QUAY_TOKEN"), "quay.io"}
	loginCmd := newExecCmd(args, cacheDir)
	if err := loginCmd.Run(); err != nil {
		return execError(args, err)
	}
	return nil
}

// PushImage pushes only the given container image to the default registry.
func PushImage(pushInfo *ImagePushInfo, buildImageDep mg.Fn) error {
	mg.SerialDeps(buildImageDep)

	// Login to container registry when running on AppSRE Jenkins.
	_, isJenkins := os.LookupEnv("JENKINS_HOME")
	_, isCI := os.LookupEnv("CI")
	if isJenkins || isCI {
		log.Println("running in CI, calling container runtime login")
		if err := quayLogin(pushInfo.Runtime, pushInfo.CacheDir); err != nil {
			return err
		}
	}

	args := []string{pushInfo.Runtime, "push"}
	if pushInfo.Runtime == string(ContainerRuntimePodman) && pushInfo.DigestFile != "" {
		args = append(args, "--digestfile="+pushInfo.DigestFile)
	}
	args = append(args, pushInfo.ImageTag)

	pushCmd := newExecCmd(args, pushInfo.CacheDir)
	if err := pushCmd.Run(); err != nil {
		return execError(args, err)
	}
	return nil
}
