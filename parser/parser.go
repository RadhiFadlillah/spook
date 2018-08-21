package parser

import (
	"bytes"
	"io/ioutil"
	fp "path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-spook/spook/model"
	"github.com/sirupsen/logrus"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// ParsePosts parse all posts inside the post directory
func ParsePosts(config model.Config) (posts []model.Post, categories []model.Group, tags []model.Group, err error) {
	logrus.Infoln("Start parsing blog posts")

	// Scan and parse all posts
	dirItems, err := ioutil.ReadDir("post")
	if err != nil {
		return nil, nil, nil, err
	}

	posts = []model.Post{}
	mapTag := map[string]int{}
	mapCategory := map[string]int{}

	for _, item := range dirItems {
		if !item.IsDir() {
			continue
		}

		// Open and read index file
		itemDir := fp.Join("post", item.Name())
		content, err := readIndexFile(itemDir)
		if err != nil {
			logrus.Errorf("Unable to read index file from %s, skipped", item.Name())
			continue
		}

		// Split metadata and content
		post := model.Post{}
		content, err = readMetadata(content, &post)
		if err != nil {
			logrus.Errorf("Unable to parse metadata from %s, skipped", item.Name())
			continue
		}

		// Make sure title is not empty
		if post.Title == "" {
			logrus.Errorf("Title is not defined in %s, skipped", item.Name())
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

		// Set post's path
		post.Path = fp.Join("/", itemDir)

		// Get post's thumbnail
		thumbnailName := getThumbnailFile(itemDir)
		if thumbnailName != "" {
			post.Thumbnail = fp.Join(post.Path, thumbnailName)
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
		posts = append(posts, post)

		category := strings.TrimSpace(post.Category)
		mapCategory[category]++

		for _, tag := range post.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			mapTag[tag]++
		}
	}

	// Convert map category and tag to slice
	categories = []model.Group{}
	for category, n := range mapCategory {
		pathName := category
		if pathName == "" {
			pathName = "uncategorized"
		}

		categories = append(categories, model.Group{
			Name:   category,
			Path:   fp.Join("/", "category", pathName),
			NPosts: n,
		})
	}

	tags = []model.Group{}
	for tag, n := range mapTag {
		tags = append(tags, model.Group{
			Name:   tag,
			Path:   fp.Join("/", "tag", tag),
			NPosts: n,
		})
	}

	// Sort list category, tag and post
	logrus.Println("Sorting posts")
	sort.Slice(posts, func(i int, j int) bool {
		iTime, _ := time.Parse("2006-01-02 15:04:05 -0700", posts[i].UpdatedAt)
		jTime, _ := time.Parse("2006-01-02 15:04:05 -0700", posts[j].UpdatedAt)
		return iTime.After(jTime)
	})

	logrus.Println("Sorting categories")
	sort.Slice(categories, func(i int, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	logrus.Println("Sorting tags")
	sort.Slice(tags, func(i int, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	// Finished
	logrus.Println("Finished parsing all posts")
	return posts, categories, tags, nil
}

// ParsePages parse all pages inside the page directory
func ParsePages(config model.Config) (pages []model.Page, err error) {
	logrus.Infoln("Start parsing pages")

	// Scan and parse all pages
	dirItems, err := ioutil.ReadDir("page")
	if err != nil {
		return nil, err
	}

	pages = []model.Page{}
	for _, item := range dirItems {
		// Open and read index file
		itemDir := fp.Join("page", item.Name())
		content, err := readIndexFile(itemDir)
		if err != nil {
			logrus.Errorf("Unable to read index file from %s, skipped", item.Name())
			continue
		}

		// Split metadata and content
		page := model.Page{}
		content, err = readMetadata(content, &page)
		if err != nil {
			logrus.Errorf("Unable to parse metadata from %s, skipped", item.Name())
			continue
		}

		// Make sure title is not empty
		if page.Title == "" {
			logrus.Errorf("Title is not defined in %s, skipped", item.Name())
			continue
		}

		// Set page's path
		page.Path = fp.Join("/", itemDir)

		// Get page's thumbnail
		thumbnailName := getThumbnailFile(itemDir)
		if thumbnailName != "" {
			page.Thumbnail = fp.Join(page.Path, thumbnailName)
		}

		// If it doesn't have any excerpt, pick the first paragraph
		if page.Excerpt == "" {
			html := blackfriday.Run(content)
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
			if err == nil {
				p := doc.Find("p").First().Text()
				page.Excerpt = strings.Join(strings.Fields(p), " ")
			}
		}

		// Save parse result
		pages = append(pages, page)
	}

	// Sort list page
	logrus.Println("Sorting pages")
	sort.Slice(pages, func(i int, j int) bool {
		return pages[i].Title < pages[j].Title
	})

	// Finished
	logrus.Println("Finished parsing all pages")
	return pages, nil
}
