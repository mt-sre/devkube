package dev

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"log"
	"os"
	"os/exec"
)

type ImageBuildInfo struct {
	ImageTag      string
	CacheDir      string
	ContainerFile string
	ContextDir    string
	Runtime       string
}

type ImagePushInfo struct {
	DigestFile string
	buildInfo  *ImageBuildInfo
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// BuildImage is a generic image build function,
// requires the binaries to be built beforehand
func BuildImage(buildInfo *ImageBuildInfo, populateCache func(), deps []interface{}) {
	if len(deps) > 0 {
		mg.Deps(deps...)
	}

	if populateCache != nil {
		populateCache()
	}

	buildCmd := []string{buildInfo.Runtime, "build", "-t", buildInfo.ImageTag}
	if buildInfo.ContainerFile != "" {
		buildCmd = append(buildCmd, "-f", buildInfo.ContainerFile)
	}
	buildCmd = append(buildCmd, buildInfo.ContextDir)

	cmds := [][]string{
		buildCmd,
		{buildInfo.Runtime, "image", "save", "-o", buildInfo.CacheDir + ".tar", buildInfo.ImageTag},
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

// PushImage builds and pushes only the given container image to the default registry.
func PushImage(pushInfo *ImagePushInfo, populateCache func(), deps []interface{}) {
	mg.SerialDeps(mg.F(BuildImage, pushInfo.buildInfo, populateCache, deps))

	// Login to container registry when running on AppSRE Jenkins.
	_, isJenkins := os.LookupEnv("JENKINS_HOME")
	_, isCI := os.LookupEnv("CI")
	if isJenkins || isCI {
		log.Println("running in CI, calling container runtime login")
		args := []string{"login", "-u=" + os.Getenv("QUAY_USER"), "-p=" + os.Getenv("QUAY_TOKEN"), "quay.io"}
		if err := sh.Run(pushInfo.buildInfo.Runtime, args...); err != nil {
			panic(fmt.Errorf("registry login: %w", err))
		}
	}

	args := []string{"push"}
	if pushInfo.buildInfo.Runtime == string(ContainerRuntimePodman) {
		args = append(args, "--digestfile="+pushInfo.DigestFile)
	}
	args = append(args, pushInfo.buildInfo.ImageTag)

	if err := sh.Run(pushInfo.buildInfo.Runtime, args...); err != nil {
		panic(fmt.Errorf("pushing image: %w", err))
	}
}
