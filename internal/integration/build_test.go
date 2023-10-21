package integration

/*
import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mt-sre/devkube/devbuild"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	projectRoot string
	cacheDir    string
	testDataDir string
)

func init() {
	dir, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	projectRoot = dir
	cacheDir = filepath.Join(projectRoot, ".cache")
	testDataDir = filepath.Join(projectRoot, "integration/test-data")
}

func buildBinary() error {
	args := []string{"build", filepath.Join(testDataDir, "test-stub/main.go")}
	cmd := exec.Command("go", args...)
	cmd.Dir = testDataDir
	return cmd.Run()
}

func cleanCache(cache string) error {
	if err := os.RemoveAll(cache); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting cache: %w", err)
	}
	if err := os.Remove(cache + ".tar"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting image cache: %w", err)
	}
	if err := os.MkdirAll(cache, os.ModePerm); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}
	return nil
}

func populateCache(cache string, files ...string) error {
	for _, f := range files {
		if err := sh.Copy(filepath.Join(cache, filepath.Base(f)), f); err != nil {
			return fmt.Errorf("copying %s: %w", f, err)
		}
	}
	return nil
}

func detectImage(tagPattern string) (bool, error) {
	out, err := exec.Command(string(runtime), "images").Output() //nolint:gosec
	if err != nil {
		return false, err
	}
	return regexp.Match(tagPattern, out)
}

func TestBuildImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cache := filepath.Join(cacheDir, "test-stub")

	deps := []any{
		mg.F(buildBinary),
		mg.F(cleanCache, cache),
		mg.F(populateCache, cache,
			filepath.Join(testDataDir, "main"),
			filepath.Join(testDataDir, "passwd"),
			filepath.Join(testDataDir, "test-stub.Containerfile")),
	}

	buildInfo := devbuild.ImageBuildInfo{
		ImageTag:      "test-stub",
		CacheDir:      cache,
		ContainerFile: "test-stub.Containerfile",
		ContextDir:    ".",
	}

	require.NoError(t, devbuild.BuildImage(ctx, &buildInfo, deps))

	match, err := detectImage("/test-stub")
	require.NoError(t, err)
	assert.True(t, match)
}

func TestBuildPackage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cache := filepath.Join(cacheDir, "test-stub-package")
	testPackageDir := filepath.Join(testDataDir, "test-stub-package")

	deps := []any{
		mg.F(cleanCache, cache),
		mg.F(populateCache, cache,
			filepath.Join(testPackageDir, "manifest.yaml"),
			filepath.Join(testPackageDir, "deployment.yaml.gotmpl"),
			filepath.Join(testPackageDir, "namespace.template.yaml.gotmpl")),
	}

	buildInfo := devbuild.PackageBuildInfo{
		ImageTag:   "test-stub-package",
		CacheDir:   cache,
		SourcePath: cache,
		OutputPath: cache + ".tar",
	}

	require.NoError(t, devbuild.BuildPackage(ctx, &buildInfo, deps))

	match, err := detectImage("/test-stub-package")
	require.NoError(t, err)
	assert.True(t, match)
}
*/
