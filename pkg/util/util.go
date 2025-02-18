/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"io"
	"os"
	"path"

	"path/filepath"

	"go.uber.org/multierr"
)

// StringSliceContains --.
func StringSliceContains(slice []string, items []string) bool {
	for i := 0; i < len(items); i++ {
		if !StringSliceExists(slice, items[i]) {
			return false
		}
	}

	return true
}

// StringSliceExists --.
func StringSliceExists(slice []string, item string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}

	return false
}

// StringSliceUniqueAdd append the given item if not already present in the slice.
func StringSliceUniqueAdd(slice *[]string, item string) bool {
	if slice == nil {
		newSlice := make([]string, 0)
		slice = &newSlice
	}
	for _, i := range *slice {
		if i == item {
			return false
		}
	}

	*slice = append(*slice, item)

	return true
}

// ReadFile a safe wrapper of os.ReadFile.
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(filename))
}

// WithFile a safe wrapper to process a file.
func WithFile(name string, flag int, perm os.FileMode, consumer func(out io.Writer) error) error {
	// #nosec G304
	file, err := os.OpenFile(filepath.Clean(name), flag, perm)
	if err == nil {
		err = consumer(file)
	}

	return Close(err, file)
}

func Close(err error, closer io.Closer) error {
	return multierr.Append(err, closer.Close())
}

// GetInWorkingDir --.
func GetInWorkingDir(dir string) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	outputDir := path.Join(workingDir, dir)
	_, err = os.Stat(outputDir)

	return outputDir, err
}

// CreateInWorkingDir --.
func CreateInWorkingDir(dir string) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	newDir := path.Join(workingDir, dir)
	err = CreateIfNotExists(newDir)
	return newDir, err
}

// RemoveFromWorkingDir --.
func RemoveFromWorkingDir(dir string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	toDelete := path.Join(workingDir, dir)
	if _, err := os.Stat(dir); err == nil {
		err = os.RemoveAll(toDelete)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

// CreateIfNotExists --.
func CreateIfNotExists(dir string) error {
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
