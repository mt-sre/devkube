// Package devos provides testable interfacing to the [os] and [os/exec] package.
package devos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Loads kubernets objects from all .yaml files in the given folder.
// Does not recurse into subfolders.
// Preserves lexical file order.
func UnstructuredFromFolder(src FS, paths ...string) ([]*unstructured.Unstructured, error) {
	src = RealFSIfUnset(src)
	var objects []*unstructured.Unstructured
	for _, path := range paths {
		files, err := src.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if filepath.Ext(file.Name()) != ".yaml" {
				continue
			}

			objs, err := UnstructuredFromFiles(src, filepath.Join(path, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("loading kubernetes objects from file %q: %w", file, err)
			}
			objects = append(objects, objs...)
		}
	}
	return objects, nil
}

func UnstructuredFromHTTP(ctx context.Context, cli *http.Client, urls ...string) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		resp, err := cli.Do(req)
		if err != nil {
			return nil, fmt.Errorf("getting %q: %w", url, err)
		}

		content := bytes.Buffer{}
		_, rErr := content.ReadFrom(resp.Body)
		cErr := resp.Body.Close()
		switch {
		case rErr != nil:
			return nil, fmt.Errorf("reading response %q: %w", url, err)
		case cErr != nil:
			return nil, fmt.Errorf("close response %q: %w", url, err)
		}

		objs, err := UnstructuredFromBytes(content.Bytes())
		if err != nil {
			return nil, fmt.Errorf("loading objects from %q: %w", url, err)
		}

		objects = append(objects, objs...)
	}

	return objects, nil
}

// Loads kubernetes objects from the given file.
func UnstructuredFromFiles(src FS, paths ...string) ([]*unstructured.Unstructured, error) {
	src = RealFSIfUnset(src)

	data := make([]byte, 0, len(paths))
	for _, path := range paths {
		fileYaml, err := src.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		data = append(data, fileYaml...)
	}

	return UnstructuredFromBytes(data)
}

// Loads kubernetes objects from given bytes.
// A single file may contain multiple objects separated by "---\n".
func UnstructuredFromBytes(fileData ...[]byte) ([]*unstructured.Unstructured, error) {
	res := make([]*unstructured.Unstructured, 0, len(fileData))
	for _, data := range fileData {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		for {
			var obj unstructured.Unstructured
			err := dec.Decode(&obj)
			switch {
			case err == nil:
				res = append(res, &obj)
				continue
			case errors.Is(err, io.EOF):
			default:
				return nil, fmt.Errorf("unmarshalling yaml document: %w", err)
			}
			break
		}
	}

	return res, nil
}

func ObjectsFromUnstructured(in []*unstructured.Unstructured) []client.Object {
	out := make([]client.Object, len(in))
	for i := range out {
		out[i] = in[i]
	}

	return out
}
