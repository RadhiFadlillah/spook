package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
)

// isEmpty checks if a directory is empty or not.
func isEmpty(dirPath string) bool {
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

// dirExists returns true if directory in specified path is exist.
func dirExists(path string) bool {
	if f, err := os.Stat(path); err == nil && f.IsDir() {
		return true
	}

	return false
}

// createFile creates empty file in specified path
func createFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return f.Close()
}

// createDirName creates a unique dir name with limited character count
func createDirName(name string, dstDir string, charLimit int) string {
	// Create dir name in lowercase
	dirName := ""
	for _, word := range strings.Fields(name) {
		dirName += strings.ToLower(word) + "-"
		if len(dirName) >= charLimit {
			break
		}
	}
	dirName = strings.TrimSuffix(dirName, "-")

	// Make sure it's unique compared to its sibling in dst dir
	for {
		dirPath := fp.Join(dstDir, dirName)
		if dirExists(dirPath) {
			dirName += "-1"
			continue
		}

		break
	}

	return dirName
}

// openConfigFile opens config file in current working directory.
// If needed, it will also check if theme specified in config file.
func openConfigFile(checkTheme bool) (model.Config, error) {
	config := model.Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		return model.Config{}, err
	}

	if checkTheme && config.Theme == "" {
		return model.Config{}, fmt.Errorf("no theme specified in config file")
	}

	return config, nil
}

func copyFile(src, dst string) error {
	src = fp.Clean(src)
	dst = fp.Clean(dst)

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create target file
	err = os.MkdirAll(fp.Dir(dst), os.ModePerm)
	if err != nil {
		return err
	}

	dstFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	dstFile, err := os.OpenFile(dst, dstFlag, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

func copyDir(src, dst string, excludedFiles ...string) error {
	src = fp.Clean(src)
	dst = fp.Clean(dst)

	// Make sure src is directory
	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !si.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	// Remove target directory, then recreate it
	err = os.RemoveAll(dst)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}

	// Copy each file and subdirectories
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := fp.Join(src, entry.Name())
		dstPath := fp.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath, excludedFiles...)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			// Skip excluded files
			isExcluded := false
			for _, excluded := range excludedFiles {
				if entry.Name() == excluded {
					isExcluded = true
					break
				}
			}

			if isExcluded {
				continue
			}

			// Copy file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removeDirContents(dirPath string) error {
	dirItems, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, dirItem := range dirItems {
		// Skip hidden file or directory (like .git)
		if strings.HasPrefix(dirItem.Name(), ".") {
			continue
		}

		// Skip CNAME as well, since it's used for Github pages
		if strings.ToLower(dirItem.Name()) == "cname" {
			continue
		}

		dirItemPath := fp.Join(dirPath, dirItem.Name())
		if dirItem.IsDir() {
			err = removeDirContents(dirItemPath)
			if err != nil {
				return err
			}
			continue
		}

		err = os.Remove(dirItemPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
