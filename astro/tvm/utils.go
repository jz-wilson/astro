package tvm

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// downloadFile will download the specified file to the specified path.
func downloadFile(url string, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func processFileContents(f *zip.File, destDir string) error {
	fh, err := f.Open()
	if err != nil {
		return err
	}
	defer func(fh io.ReadCloser) {
		err := fh.Close()
		if err != nil {

		}
	}(fh)

	path := filepath.Join(destDir, f.Name)

	if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path in zip: %s", path)
	}

	if f.FileInfo().IsDir() {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(out, fh)

		err = out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(zipfilePath string, destDir string) error {
	r, err := zip.OpenReader(zipfilePath)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {

		}
	}(r)

	for _, f := range r.File {
		err = processFileContents(f, destDir)
		if err != nil {
			return err
		}
	}
	return nil
}
