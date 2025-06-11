package helper

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetProjectRoot_WithGoMod(t *testing.T) {
	tempDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testmodule"), 0644)
	require.NoError(t, err)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	root, err := helper.GetProjectRoot()
	require.NoError(t, err)
	require.Equal(t, tempDir, root)
}

func TestGetProjectRoot_WithGitDir(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
	require.NoError(t, err)

	// Change working directory to temp
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	root, err := helper.GetProjectRoot()
	require.NoError(t, err)
	require.Equal(t, tempDir, root)
}

func TestGetProjectRoot_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	root, err := helper.GetProjectRoot()
	require.Error(t, err)
	require.Equal(t, "", root)
	require.Contains(t, err.Error(), "project root not found")
}

func TestGetProjectRoot_FailsOnGetwd(t *testing.T) {
	original := helper.GetWdFunc
	defer func() { helper.GetWdFunc = original }()

	helper.GetWdFunc = func() (string, error) {
		return "", errors.New("simulated Getwd failure")
	}

	root, err := helper.GetProjectRoot()

	require.Error(t, err)
	require.Empty(t, root)
	require.Contains(t, err.Error(), "simulated Getwd failure")
}
