package helper

import (
	"errors"
	"os"
	"path/filepath"
)

var GetWdFunc = os.Getwd

var GetProjectRoot = func() (string, error) {
	dir, err := GetWdFunc()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("project root not found")
}
