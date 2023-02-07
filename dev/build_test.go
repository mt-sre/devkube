package dev

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
)

type buildTestCase struct {
	name      string
	buildInfo ImageBuildInfo
	buildCmd  []string
	saveCmd   []string
}

type pushTestCase struct {
	name     string
	pushInfo ImagePushInfo
	pushCmd  []string
	loginCmd []string
}

var (
	defaultBuildCase = buildTestCase{
		name: "default",
		buildInfo: ImageBuildInfo{
			ImageTag:      "test_ImageTag",
			CacheDir:      "",
			ContainerFile: "test_ContainerFile",
			ContextDir:    "test_ContextDir",
			Runtime:       "test_Runtime",
		},
		buildCmd: []string{"test_Runtime", "build", "-t", "test_ImageTag", "-f", "test_ContainerFile", "test_ContextDir"},
		saveCmd:  []string{"test_Runtime", "image", "save", "-o", ".tar", "test_ImageTag"},
	}

	noConFileBuildCase = buildTestCase{
		name: "no-container-file",
		buildInfo: ImageBuildInfo{
			ImageTag:      "test_ImageTag",
			CacheDir:      "",
			ContainerFile: "",
			ContextDir:    "test_ContextDir",
			Runtime:       "test_Runtime",
		},
		buildCmd: []string{"test_Runtime", "build", "-t", "test_ImageTag", "test_ContextDir"},
		saveCmd:  []string{"test_Runtime", "image", "save", "-o", ".tar", "test_ImageTag"},
	}

	buildTestCases = map[string]*buildTestCase{
		"default":           &defaultBuildCase,
		"no-container-file": &noConFileBuildCase,
	}

	defaultPushCase = pushTestCase{
		name: "default",
		pushInfo: ImagePushInfo{
			ImageTag:   "test_ImageTag",
			CacheDir:   "",
			Runtime:    "test_Runtime",
			DigestFile: "test_DigestFile",
		},
		pushCmd:  []string{"test_Runtime", "push", "test_ImageTag"},
		loginCmd: []string{"test_Runtime", "login", "-u=" + os.Getenv("QUAY_USER"), "-p=" + os.Getenv("QUAY_TOKEN"), "quay.io"},
	}

	podmanPushCase = pushTestCase{
		name: "podman",
		pushInfo: ImagePushInfo{
			ImageTag:   "test_ImageTag",
			CacheDir:   "",
			Runtime:    string(ContainerRuntimePodman),
			DigestFile: "test_DigestFile",
		},
		pushCmd:  []string{string(ContainerRuntimePodman), "push", "--digestfile=test_DigestFile", "test_ImageTag"},
		loginCmd: []string{string(ContainerRuntimePodman), "login", "-u=" + os.Getenv("QUAY_USER"), "-p=" + os.Getenv("QUAY_TOKEN"), "quay.io"},
	}

	pushTestCases = map[string]*pushTestCase{
		"default": &defaultPushCase,
		"podman":  &podmanPushCase,
	}

	// currentTestCase is used in TestXXXX_HelperProcess to identify which test ran it
	currentTestCase string

	// helperProcess is used by mockExecCommand to determine which helper process to run
	helperProcess string
)

const (
	buildHelper = "TestBuildImage_HelperProcess"
	pushHelper  = "TestPushImage_HelperProcess"
)

func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=" + helperProcess, "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"GO_TEST_CASE_NAME=" + currentTestCase,
	}
	return cmd
}

func TestBuildImage_HelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	tc := buildTestCases[os.Getenv("GO_TEST_CASE_NAME")]
	command := os.Args[3:]
	switch command[1] {
	case "build":
		assert.Equal(t, tc.buildCmd, command)
	case "image":
		assert.Equal(t, tc.saveCmd, command)
	default:
		t.Errorf("invalid command")
	}
	os.Exit(0)
}

func TestBuildImage(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	helperProcess = buildHelper

	for _, tc := range buildTestCases {
		currentTestCase = tc.name
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, BuildImage(&tc.buildInfo, []interface{}{}))
		})
	}
}

func TestPushImage_HelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	tc := pushTestCases[os.Getenv("GO_TEST_CASE_NAME")]
	command := os.Args[3:]
	switch command[1] {
	case "push":
		assert.Equal(t, tc.pushCmd, command)
	case "login":
		fmt.Println(command)
		assert.Equal(t, tc.loginCmd, command)
	}
	os.Exit(0)
}

func TestPushImage(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	helperProcess = pushHelper

	for _, tc := range pushTestCases {
		currentTestCase = tc.name
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, PushImage(&tc.pushInfo, mg.F(func() {})))
		})
	}
}
