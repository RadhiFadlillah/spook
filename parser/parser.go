package parser

import (
	"fmt"
	"io/ioutil"
	fp "path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-spook/spook/model"
)

// Parser is used to parse markdown files to get the posts,
// pages, categories and tags that used in the blog.
type Parser struct {
	Config  model.Config
	RootDir string
}

// ParsedPosts is the output from ParsePosts
type ParsedPosts struct {
	Posts      []model.Post
	Categories []model.Group
	Tags       []model.Group
}

// ParsePosts parse all posts inside the post directory.
// Returns all posts, also categories and tags that used in the posts.
func (ps Parser) ParsePosts() (output ParsedPosts, err error) {
	// The valid posts must be structured like this :
	// <root-dir>
	// `-- post
	//     |-- 2006-02-03-post-name-1
	//     |   |-- _index.md
	//     |   |-- _thumbnail.jpg
	//     |   |-- image.jpg
	//     |   `-- sample.txt
	//     `-- 2006-02-03-post-name-2

	// Scan and parse all posts.
	postDir := fp.Join(ps.RootDir, "post")
	dirItems, err := ioutil.ReadDir(postDir)
	if err != nil {
		return output, fmt.Errorf("failed to scan post dir: %s", err)
	}

	posts := []model.Post{}
	mapTag := map[string]int{}
	mapCategory := map[string]int{}

	for _, item := range dirItems {
		if !item.IsDir() {
			continue
		}

		// Open and read index file
		itemDir := fp.Join(postDir, item.Name())
		content, err := readIndexFile(itemDir)
		if err != nil {
			return output, fmt.Errorf("failed to read index file for %s: %s", item.Name(), err)
		}

		// Split metadata and content
		post := model.Post{}
		content, err = readMetadata(content, &post)
		if err != nil {
			return output, fmt.Errorf("failed to parse metadata for %s: %s", item.Name(), err)
		}

		// Make sure title is not empty
		if post.Title == "" {
			return output, fmt.Errorf("title is not defined in %s", item.Name())
		}

		// Make sure date time format is correct
		if post.UpdatedAt == "" {
			post.UpdatedAt = post.CreatedAt
		}

		if _, err = time.Parse("2006-01-02 15:04:05 -0700", post.CreatedAt); err != nil {
			return output, fmt.Errorf("failed to parse create time for %s: %s", item.Name(), err)
		}

		if _, err = time.Parse("2006-01-02 15:04:05 -0700", post.UpdatedAt); err != nil {
			return output, fmt.Errorf("failed to parse update time for %s: %s", item.Name(), err)
		}

		// Set post's path
		post.Path = fp.Join("/", "post", item.Name())

		// Get post's thumbnail
		thumbnailName := getThumbnailFile(itemDir)
		if thumbnailName != "" {
			post.Thumbnail = fp.Join(post.Path, thumbnailName)
		}

		// If it doesn't have any excerpt, pick the first paragraph
		if post.Excerpt == "" {
			post.Excerpt = getFirstParagraph(content)
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
	categories := []model.Group{}
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

	tags := []model.Group{}
	for tag, n := range mapTag {
		tags = append(tags, model.Group{
			Name:   tag,
			Path:   fp.Join("/", "tag", tag),
			NPosts: n,
		})
	}

	// Sort list category, tag and post
	sort.Slice(posts, func(i int, j int) bool {
		iTime, _ := time.Parse("2006-01-02 15:04:05 -0700", posts[i].UpdatedAt)
		jTime, _ := time.Parse("2006-01-02 15:04:05 -0700", posts[j].UpdatedAt)
		return iTime.After(jTime)
	})

	sort.Slice(categories, func(i int, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	sort.Slice(tags, func(i int, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	// Finished
	output = ParsedPosts{
		Posts:      posts,
		Categories: categories,
		Tags:       tags,
	}

	return output, nil
}

// ParsePages parse all pages inside the page directory
func (ps Parser) ParsePages() (pages []model.Page, err error) {
	// The valid pages must be structured like this :
	// <root-dir>
	// `-- page
	//     |-- page-1
	//     |   |-- _index.md
	//     |   `-- _thumbnail.jpg
	//     `-- page-2

	// Scan and parse all pages
	pageDir := fp.Join(ps.RootDir, "page")
	dirItems, err := ioutil.ReadDir(pageDir)
	if err != nil {
		return nil, err
	}

	pages = []model.Page{}
	for _, item := range dirItems {
		if !item.IsDir() {
			continue
		}

		// Open and read index file
		itemDir := fp.Join(pageDir, item.Name())
		content, err := readIndexFile(itemDir)
		if err != nil {
			return pages, fmt.Errorf("failed to read index file for %s: %s", item.Name(), err)
		}

		// Split metadata and content
		page := model.Page{}
		content, err = readMetadata(content, &page)
		if err != nil {
			return pages, fmt.Errorf("failed to parse metadata for %s: %s", item.Name(), err)
		}

		// Make sure title is not empty
		if page.Title == "" {
			return pages, fmt.Errorf("title is not defined in %s", item.Name())
		}

		// Set page's path
		page.Path = fp.Join("/", "page", item.Name())

		// Get page's thumbnail
		thumbnailName := getThumbnailFile(itemDir)
		if thumbnailName != "" {
			page.Thumbnail = fp.Join(page.Path, thumbnailName)
		}

		// If it doesn't have any excerpt, pick the first paragraph
		if page.Excerpt == "" {
			page.Excerpt = getFirstParagraph(content)
		}

		// Save parse result
		pages = append(pages, page)
	}

	// Sort list page
	sort.Slice(pages, func(i int, j int) bool {
		return pages[i].Title < pages[j].Title
	})

	// Finished
	return pages, nil
}
