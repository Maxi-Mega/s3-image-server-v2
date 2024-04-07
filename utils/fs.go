package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
)

// CreateDir creates a directory at the given path.
// If it already exists, its content is cleared.
// If a file exists at this path, is it overridden.
func CreateDir(dirPath string) error {
	stat, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dirPath, 0700) //nolint:wrapcheck
		}

		return err //nolint:wrapcheck
	}

	if stat.IsDir() {
		return ClearDir(dirPath)
	}

	logger.Debug("Removing file at ", dirPath)

	err = os.Remove(dirPath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return os.Mkdir(dirPath, 0700) //nolint:wrapcheck
}

// ClearDir removes all the entries from the given directory.
func ClearDir(dirPath string) error {
	logger.Trace("Clearing dir at ", dirPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	for _, entry := range entries {
		err = os.RemoveAll(filepath.Join(dirPath, entry.Name()))
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

func FormatDirName(s3Dir string) string {
	return strings.ReplaceAll(strings.Trim(s3Dir, "/"), "/", "@")
}

// FileStat executes a stat against the file at the given path.
// If the file does not exist, no error but false is returned.
func FileStat(filePath string) (stat os.FileInfo, exists bool, err error) {
	switch stat, err = os.Stat(filePath); {
	case err == nil:
		return stat, true, nil
	case os.IsNotExist(err):
		return nil, false, nil
	default:
		return nil, false, err //nolint:wrapcheck
	}
}
