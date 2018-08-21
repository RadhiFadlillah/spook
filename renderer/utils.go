package renderer

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma"
	fhtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// fileExists returns true if file in path is exist.
func fileExists(path string) bool {
	if f, err := os.Stat(path); err == nil && f.Size() > 0 {
		return true
	}

	return false
}

// readIndexFile reads content of _index.md file in specified directory
func readIndexFile(dir string) ([]byte, error) {
	indexFile, err := os.Open(fp.Join(dir, "_index.md"))
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()

	return ioutil.ReadAll(indexFile)
}

// removeMetadata removes metadata from specified content
func removeMetadata(content []byte) []byte {
	origin := content
	if !bytes.HasPrefix(content, []byte("+++\n")) {
		return origin
	}

	content = bytes.TrimPrefix(content, []byte("+++\n"))
	separatorIdx := bytes.Index(content, []byte("+++\n"))
	if separatorIdx == -1 {
		return origin
	}

	return content[separatorIdx+3:]
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

// highlighCode highlights the code in generated HTML
func highlightCode(html []byte) []byte {
	r := bytes.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return html
	}

	formatter := fhtml.New(fhtml.WithClasses())
	doc.Find("pre>code").Each(func(_ int, cd *goquery.Selection) {
		text := cd.Text()
		text = strings.TrimSuffix(text, "\n")

		lexer := lexers.Analyse(text)
		if lexer == nil {
			lexer = lexers.Fallback
		}
		lexer = chroma.Coalesce(lexer)

		iterator, err := lexer.Tokenise(nil, text)
		if err != nil {
			return
		}

		output := bytes.Buffer{}
		err = formatter.Format(&output, styles.Fallback, iterator)
		cd.SetHtml(output.String())
	})

	newHTML, err := doc.Html()
	if err != nil {
		return html
	}

	return []byte(newHTML)
}
