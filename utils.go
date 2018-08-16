package main

import (
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"
)

func fileExists(path string) bool {
	if f, err := os.Stat(path); err == nil && f.Size() > 0 {
		return true
	}

	return false
}

func getBaseTemplate(themeDir string) []string {
	items, err := ioutil.ReadDir(themeDir)
	if err != nil {
		return []string{}
	}

	templates := []string{}
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		if strings.HasSuffix(item.Name(), ".html") && strings.HasPrefix(item.Name(), "_") {
			templates = append(templates, fp.Join(themeDir, item.Name()))
		}
	}

	return templates
}
