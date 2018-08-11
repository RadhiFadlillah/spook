package main

import (
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"
	"time"
)

func createFile(path string) {
	f, _ := os.Create(path)
	defer f.Close()
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

func formatTime(strTime string, dstFormat string) string {
	dt, err := time.Parse("2006-01-02 15:04:05 -0700", strTime)
	if err != nil {
		return ""
	}

	return dt.Format(dstFormat)
}

func limitSentence(src string, n int) string {
	if src == "" {
		return src
	}

	tempSrc := src
	tempSrc = strings.Replace(tempSrc, ".", "||", -1)
	tempSrc = strings.Replace(tempSrc, "?", "||", -1)
	tempSrc = strings.Replace(tempSrc, "!", "||", -1)
	tempSrc = strings.Replace(tempSrc, "\n", "||", -1)
	tempSrc = strings.TrimSuffix(tempSrc, "||")

	sentences := strings.Split(tempSrc, "||")
	if n > len(sentences) {
		n = len(sentences)
	}

	result := strings.Join(sentences[:n], ".")
	return src[:len(result)+1]
}
