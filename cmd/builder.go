package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/go-spook/spook/model"
	"github.com/go-spook/spook/parser"
	"github.com/go-spook/spook/renderer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmdBuild = &cobra.Command{
	Use:     "build",
	Short:   "Build the static site",
	Aliases: []string{"builder"},
	Args:    cobra.NoArgs,
	Run:     cmdBuildHandler,
}

func cmdBuildHandler(cmd *cobra.Command, args []string) {
	// Make sure valid config file exists in current working dir
	config, err := openConfigFile(true)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	config.Theme = fp.Join("theme", config.Theme)

	// Parse all posts and pages
	posts, categories, tags, err := parser.ParsePosts(config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	pages, err := parser.ParsePages(config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Set destination directory
	dstDir := config.PublishDir
	if dstDir == "" {
		dstDir = "public"
	}

	// Make sure directory exists and clean it
	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	err = removeDirContents(dstDir)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Copy static directory if it exists
	if sd, err := os.Stat("static"); err == nil && sd.IsDir() {
		err = copyDir("static", fp.Join(dstDir, "static"), true)
		if err != nil {
			cError.Println("Error: Unable to copy static directory:", err)
			return
		}
	}

	// Copy theme directory
	themeItems, err := ioutil.ReadDir(config.Theme)
	if err != nil {
		cError.Println("Error: Unable to copy theme files:", err)
		return
	}

	for _, item := range themeItems {
		if !item.IsDir() {
			continue
		}

		err = copyDir(fp.Join(config.Theme, item.Name()), fp.Join(dstDir, item.Name()), true)
		if err != nil {
			cError.Println("Error: Unable to copy theme files:", err)
			return
		}
	}

	// Create renderer
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      posts,
		Tags:       tags,
		Categories: categories,
		Minimize:   true,
	}

	// Build frontpage
	logrus.Println("Building front page")
	err = buildFrontPage(rd, dstDir)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Build list of posts
	logrus.Println("Building list of posts")
	postsDir := fp.Join(dstDir, "posts")
	err = buildList(rd, postsDir, renderer.DEFAULT, "")
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Build list of posts by category
	logrus.Println("Building list of posts by category")
	for _, category := range categories {
		if category.Name == "" {
			category.Name = "uncategorized"
		}
		categoryDir := fp.Join(dstDir, "category", category.Name)

		err = buildList(rd, categoryDir, renderer.CATEGORY, category.Name)
		if err != nil {
			cError.Println("Error:", err)
			return
		}
	}

	// Build list of posts by tag
	logrus.Println("Building list of posts by tag")
	for _, tag := range tags {
		tagDir := fp.Join(dstDir, "tag", tag.Name)
		err = buildList(rd, tagDir, renderer.TAG, tag.Name)
		if err != nil {
			cError.Println("Error:", err)
			return
		}
	}

	// Build pages
	logrus.Println("Building pages")
	err = buildPages(rd, dstDir, pages)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Build posts
	logrus.Println("Building posts")
	err = buildPosts(rd, dstDir, posts)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
}

func buildFrontPage(rd renderer.Renderer, dstDir string) error {
	frontPage, err := os.Create(fp.Join(dstDir, "index.html"))
	if err != nil {
		return fmt.Errorf("Unable to create index file for front page: %v", err)
	}

	err = rd.RenderFrontPage(frontPage)
	if err != nil {
		return fmt.Errorf("Unable to render front page: %v", err)
	}

	return nil
}

func buildList(rd renderer.Renderer, dstDir string, listType renderer.ListType, groupName string) error {
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to build list: %v", err)
	}

	for i := 0; ; i++ {
		fileName := "index.html"
		if i > 0 {
			fileName = fmt.Sprintf("%d.html", i)
		}
		fileName = fp.Join(dstDir, fileName)

		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("Unable to create file for list of posts: %v", err)
		}

		nPosts, err := rd.RenderList(listType, "", i, f)
		if err != nil {
			f.Close()
			return fmt.Errorf("Unable to build list of posts: %v", err)
		}

		f.Close()

		if nPosts == -1 {
			os.Remove(fileName)
			break
		}
	}

	return nil
}

func buildPages(rd renderer.Renderer, dstDir string, pages []model.Page) error {
	for _, page := range pages {
		page.Path = strings.TrimPrefix(page.Path, "/")
		err := copyDir(page.Path, fp.Join(dstDir, page.Path), true, "_index.md")
		if err != nil {
			return fmt.Errorf("Unable to copy file for %s: %v, skipped", page.Path, err)
		}

		f, err := os.Create(fp.Join(dstDir, page.Path, "index.html"))
		if err != nil {
			return fmt.Errorf("Unable to create index file for %s: %v, skipped", page.Path, err)
		}

		err = rd.RenderPage(page, f)
		if err != nil {
			f.Close()
			return fmt.Errorf("Unable to build %s: %v, skipped", page.Path, err)
		}

		f.Close()
	}

	return nil
}

func buildPosts(rd renderer.Renderer, dstDir string, posts []model.Post) error {
	for i, post := range posts {
		post.Path = strings.TrimPrefix(post.Path, "/")
		err := copyDir(post.Path, fp.Join(dstDir, post.Path), true, "_index.md")
		if err != nil {
			return fmt.Errorf("Unable to copy file for %s: %v, skipped", post.Path, err)
		}

		f, err := os.Create(fp.Join(dstDir, post.Path, "index.html"))
		if err != nil {
			return fmt.Errorf("Unable to create index file for %s: %v, skipped", post.Path, err)
		}

		newerPost := model.Post{}
		olderPost := model.Post{}

		if i > 0 {
			newerPost = posts[i-1]
		}

		if i < len(posts)-1 {
			olderPost = posts[i+1]
		}

		err = rd.RenderPost(post, olderPost, newerPost, f)
		if err != nil {
			f.Close()
			return fmt.Errorf("Unable to build %s: %v, skipped", post.Path, err)
		}

		f.Close()
	}

	return nil
}
