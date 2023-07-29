/*
 *  Copyright (c) 2019 Uber Technologies, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tvm

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZipSlip tests to ensure we aren't being exploited by zip files with
// "../" in the file paths.
func TestZipSlip(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "astro-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tmpDir)

	// Create zip path
	tmpZipFileName := filepath.Join(tmpDir, "1/bad.zip")
	err = os.MkdirAll(filepath.Dir(tmpZipFileName), 0755)
	require.NoError(t, err)

	// Create zip file
	tmpZipFile, err := os.Create(tmpZipFileName)
	require.NoError(t, err)
	defer func(tmpZipFile *os.File) {
		err := tmpZipFile.Close()
		if err != nil {

		}
	}(tmpZipFile)

	zipWriter := zip.NewWriter(tmpZipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {

		}
	}(zipWriter)

	// Add some files
	readmeFile, err := zipWriter.Create("README.txt")
	require.NoError(t, err)
	_, err = readmeFile.Write([]byte("This is a zip file for testing."))
	require.NoError(t, err)

	// Add a naughty file
	badFile, err := zipWriter.Create("../naughty.txt")
	require.NoError(t, err)
	_, err = badFile.Write([]byte("This file should never be extracted."))
	require.NoError(t, err)

	// Write zip
	require.NoError(t, zipWriter.Close())

	// Test that extracting this zip file causes an error
	tmpDir, err = os.MkdirTemp("", "astro-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tmpDir)

	err = unzip(tmpZipFile.Name(), tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path in zip")
}
