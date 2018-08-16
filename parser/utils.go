package parser

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// readIndexFile reads content of _index.md file in specified directory
func readIndexFile(dir string) ([]byte, error) {
	indexFile, err := os.Open(fp.Join(dir, "_index.md"))
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()

	return ioutil.ReadAll(indexFile)
}

// readMetadata fetch metadata from specified content, put it to the
// specified destination, and returns final content without the metadata.
func readMetadata(content []byte, dst interface{}) ([]byte, error) {
	// Separate metadata and content
	if !bytes.HasPrefix(content, []byte("+++\n")) {
		return nil, fmt.Errorf("Content is not started with metadata")
	}

	content = bytes.TrimPrefix(content, []byte("+++\n"))
	separatorIdx := bytes.Index(content, []byte("+++\n"))
	if separatorIdx == -1 {
		return nil, fmt.Errorf("Content is not started with metadata")
	}

	metadata := content[:separatorIdx]
	content = content[separatorIdx+3:]

	// Parse metadata
	_, err := toml.Decode(string(metadata), dst)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse metadata")
	}

	return content, nil
}

// getThumbnailFile fetch thumbnail file in specified directory
func getThumbnailFile(dir string) string {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		if strings.HasPrefix(item.Name(), "_thumbnail.") {
			imgPath := fp.Join(dir, item.Name())
			if isImageFile(imgPath) {
				return item.Name()
			}
		}
	}

	return ""
}

// isImageFile check file's header and see if it is image
func isImageFile(path string) bool {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read the first 512 bytes
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return false
	}

	// Get mime type
	mimeType := http.DetectContentType(buffer)
	return strings.HasPrefix(mimeType, "image/")
}
