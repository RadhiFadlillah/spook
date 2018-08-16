package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/RadhiFadlillah/spook/model"
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

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
