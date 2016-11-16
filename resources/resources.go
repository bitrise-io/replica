package resources

import (
	"fmt"
	"os"

	"strings"

	"path/filepath"

	"io/ioutil"

	rice "github.com/GeertJohan/go.rice"
)

// GetResourcesBox ...
func GetResourcesBox() (*rice.Box, error) {
	return rice.FindBox("data")
}

// UncompressDirectory ...
func UncompressDirectory(resourceDirPath, targetDirPath string) error {
	dataBox, err := GetResourcesBox()
	if err != nil {
		return fmt.Errorf("Failed to open embedded resource, error: %s", err)
	}

	if err := os.MkdirAll(targetDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create target directory (path: %s), error: %s", targetDirPath, err)
	}

	return dataBox.Walk(resourceDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Failed to uncompress file (resource path: %s), error: %s", path, err)
		}
		if resourceDirPath == path {
			// skip
			return nil
		}

		relativePath := strings.TrimPrefix(path, resourceDirPath+"/")
		targetPath := filepath.Join(targetDirPath, relativePath)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("Failed to create directory (path: %s), error: %s", targetPath, err)
			}
		} else {
			contBytes, err := dataBox.Bytes(path)
			if err != nil {
				return fmt.Errorf("Failed to read embedded resource (path: %s), error: %s", path, err)
			}
			if err := ioutil.WriteFile(targetPath, contBytes, 0755); err != nil {
				return fmt.Errorf("Failed to write resource (path: %s) into file (path: %s), error: %s", path, targetPath, err)
			}
		}

		return nil
	})
}
