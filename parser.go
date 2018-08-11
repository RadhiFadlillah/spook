package main

import (
	"bytes"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// ParseAllPosts parse all posts inside the post directory
func ParseAllPosts(config Config, postDir string) (ParsedAllPosts, error) {
	logrus.Infoln("Start parsing blog posts")

	// Scan and parse all posts
	dirItems, err := ioutil.ReadDir(postDir)
	if err != nil {
		return ParsedAllPosts{}, err
	}

	listPost := []Post{}
	mapTag := map[string]int{}
	mapCategory := map[string]int{}

	for _, item := range dirItems {
		if !item.IsDir() {
			continue
		}

		// Open index file
		itemDir := fp.Join(postDir, item.Name())
		indexFile, err := os.Open(fp.Join(itemDir, "_index.md"))
		if err != nil {
			logrus.Errorf("Unable to open index file from %s, skipped", item.Name())
			continue
		}

		// Read content
		content, err := ioutil.ReadAll(indexFile)
		if err != nil {
			logrus.Errorf("Unable to read index file from %s, skipped", item.Name())
			continue
		}
		indexFile.Close()

		// Separate metadata and content
		if !bytes.HasPrefix(content, []byte("+++\n")) {
			logrus.Errorf("Unable to read metadata from %s, skipped", item.Name())
			continue
		}

		content = bytes.TrimPrefix(content, []byte("+++\n"))
		separatorIdx := bytes.Index(content, []byte("+++\n"))

		metadata := content[:separatorIdx]
		content = content[separatorIdx+3:]

		// Parse metadata
		post := Post{}
		_, err = toml.Decode(string(metadata), &post)
		if err != nil {
			logrus.Errorf("Unable to parse metadata from %s, skipped", item.Name())
			continue
		}

		// Make sure date time format is correct
		if post.UpdatedAt == "" {
			post.UpdatedAt = post.CreatedAt
		}

		if _, err = time.Parse("2006-01-02 15:04:05 -0700", post.CreatedAt); err != nil {
			logrus.Errorf("Unable to parse date time from %s, skipped", item.Name())
			continue
		}

		if _, err = time.Parse("2006-01-02 15:04:05 -0700", post.UpdatedAt); err != nil {
			logrus.Errorf("Unable to parse date time from %s, skipped", item.Name())
			continue
		}

		// Set post URL and author
		post.URL = itemDir
		if post.Author == "" {
			post.Author = config.Owner
		}

		// If it doesn't have any excerpt, pick the first paragraph
		if post.Excerpt == "" {
			html := blackfriday.Run(content)
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
			if err == nil {
				p := doc.Find("p").First().Text()
				post.Excerpt = strings.Join(strings.Fields(p), " ")
			}
		}

		// Save parse result
		listPost = append(listPost, post)
		mapCategory[post.Category]++
		for _, tag := range post.Tags {
			mapTag[tag]++
		}
	}

	// Convert map category and tag to slice
	listCategory := []Group{}
	for category, n := range mapCategory {
		listCategory = append(listCategory, Group{
			Name:   category,
			NPosts: n,
		})
	}

	listTag := []Group{}
	for tag, n := range mapTag {
		listTag = append(listTag, Group{
			Name:   tag,
			NPosts: n,
		})
	}

	// Sort list category, tag and post
	logrus.Println("Sorting list of categories")
	sort.Slice(listCategory, func(i int, j int) bool {
		return listCategory[i].Name < listCategory[j].Name
	})

	logrus.Println("Sorting list of tags")
	sort.Slice(listTag, func(i int, j int) bool {
		return listTag[i].Name < listTag[j].Name
	})

	logrus.Println("Sorting list of posts")
	sort.Slice(listPost, func(i int, j int) bool {
		iTime, _ := time.Parse("2006-01-02 15:04:05 -0700", listPost[i].UpdatedAt)
		jTime, _ := time.Parse("2006-01-02 15:04:05 -0700", listPost[j].UpdatedAt)
		return iTime.After(jTime)
	})

	// Finished
	logrus.Println("Finished parsing all posts")
	return ParsedAllPosts{
		Tags:       listTag,
		Categories: listCategory,
		Posts:      listPost,
	}, nil
}
