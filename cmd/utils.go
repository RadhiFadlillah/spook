package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
)

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

func createFile(path string) {
	f, _ := os.Create(path)
	defer f.Close()
}

func createDirName(name string, dst string, wordLimit int) string {
	dirPath := ""

	// Format name and limit words
	for _, word := range strings.Fields(name) {
		dirPath += strings.ToLower(word) + "-"
		if len(dirPath) >= wordLimit {
			break
		}
	}

	// Make sure it's unique
	dirPath = fp.Join(dst, dirPath[:len(dirPath)-1])
	for {
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			dirPath += "-1"
			continue
		} else {
			break
		}
	}

	return dirPath
}

func openConfigFile(checkTheme bool) (model.Config, error) {
	config := model.Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		return model.Config{}, err
	}

	if config.BaseURL == "" {
		return model.Config{}, fmt.Errorf("No base URL set in configuration file")
	}

	if _, err = url.ParseRequestURI(config.BaseURL); err != nil {
		return model.Config{}, fmt.Errorf("Base URL must be an absolute URL path")
	}

	if checkTheme && config.Theme == "" {
		return model.Config{}, fmt.Errorf("No theme specified in configuration file")
	}

	return config, nil
}

func copyFile(src, dst string, force bool) error {
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

	dstFlag := os.O_RDWR | os.O_CREATE
	if force {
		dstFlag = dstFlag | os.O_TRUNC
	}

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

func copyDir(src, dst string, force bool, excludedFiles ...string) error {
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

	// Make sure destination is not exists (unless forced)
	if force {
		err = os.RemoveAll(dst)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		return fmt.Errorf("Destination already exists")
	}

	// Create destination directory
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := fp.Join(src, entry.Name())
		dstPath := fp.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath, force, excludedFiles...)
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
			err = copyFile(srcPath, dstPath, force)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
