package cmd

import (
	"io"
	"os"
)

// dirIsEmpty checks if a directory is empty or not.
func dirIsEmpty(dirPath string) bool {
	dir, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer dir.Close()

	_, err = dir.Readdirnames(1)
	if err != io.EOF {
		return false
	}

	return true
}
