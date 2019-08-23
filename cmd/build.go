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

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "build",
		Short:   "Build the static site",
		Aliases: []string{"builder"},
		Args:    cobra.NoArgs,
		Run:     buildHandler,
	}

	cmd.Flags().StringP("output", "o", "public", "path to output directory")

	return cmd
}

func buildHandler(cmd *cobra.Command, args []string) {
	// Make sure valid config file exists in current working dir
	config, err := openConfigFile(true)
	if err != nil {
		cError.Println("Failed to open config file:", err)
		return
	}

	// Get working dir
	rootDir, err := os.Getwd()
	if err != nil {
		cError.Println("Failed to get working dir:", err)
		return
	}

	// Get output directory and clean it
	outputDir, _ := cmd.Flags().GetString("output")
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		cError.Println("Failed to create output dir:", err)
		return
	}

	err = removeDirContents(outputDir)
	if err != nil {
		cError.Println("Failed to clean output dir:", err)
		return
	}

	// Copy static directory if it exists
	staticDir := fp.Join(rootDir, "static")
	dstStaticDir := fp.Join(outputDir, "static")

	if dirExists(staticDir) {
		err = copyDir(staticDir, dstStaticDir)
		if err != nil {
			cError.Println("Failed to copy static directory:", err)
			return
		}
	}

	// Copy theme directory
	themeDir := fp.Join(rootDir, "theme", config.Theme)
	themeItems, err := ioutil.ReadDir(themeDir)
	if err != nil {
		cError.Println("Failed to read theme dir:", err)
		return
	}

	for _, item := range themeItems {
		if !item.IsDir() {
			continue
		}

		srcDir := fp.Join(themeDir, item.Name())
		dstDir := fp.Join(outputDir, item.Name())

		err = copyDir(srcDir, dstDir)
		if err != nil {
			cError.Println("Failed to copy theme files:", err)
			return
		}
	}

	// Parse all posts and pages
	psr := parser.Parser{
		Config:  config,
		RootDir: rootDir,
	}

	parsedPosts, err := psr.ParsePosts()
	if err != nil {
		cError.Println("Failed to parse posts:", err)
		return
	}

	pages, err := psr.ParsePages()
	if err != nil {
		cError.Println("Failed to parse pages:", err)
		return
	}

	// Create renderer
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      parsedPosts.Posts,
		Tags:       parsedPosts.Tags,
		Categories: parsedPosts.Categories,
		RootDir:    rootDir,
		Minimize:   true,
	}

	// Build frontpage
	logrus.Println("Building front page")
	err = buildFrontPage(rd, outputDir)
	if err != nil {
		cError.Println("Failed to build front page:", err)
		return
	}

	// Build list of posts
	logrus.Println("Building list of posts")
	postsDir := fp.Join(outputDir, "posts")
	err = buildList(rd, postsDir, renderer.DEFAULT, "")
	if err != nil {
		cError.Println("Failed to build main list:", err)
		return
	}

	// Build list of posts by category
	logrus.Println("Building list of posts by category")
	for _, category := range parsedPosts.Categories {
		if category.Name == "" {
			category.Name = "uncategorized"
		}

		categoryDir := fp.Join(outputDir, "category", category.Name)

		err = buildList(rd, categoryDir, renderer.CATEGORY, category.Name)
		if err != nil {
			cError.Printf("Failed to build list category \"%s\": %v\n:", category.Name, err)
			return
		}
	}

	// Build list of posts by tag
	logrus.Println("Building list of posts by tag")
	for _, tag := range parsedPosts.Tags {
		tagDir := fp.Join(outputDir, "tag", tag.Name)

		err = buildList(rd, tagDir, renderer.TAG, tag.Name)
		if err != nil {
			cError.Printf("Failed to build list tag \"%s\": %v\n:", tag.Name, err)
			return
		}
	}

	// Build pages
	logrus.Println("Building pages")
	err = buildPages(rd, outputDir, pages)
	if err != nil {
		cError.Println("Failed to build pages:", err)
		return
	}

	// Build posts
	logrus.Println("Building posts")
	err = buildPosts(rd, outputDir, parsedPosts.Posts)
	if err != nil {
		cError.Println("Failed to build posts:", err)
		return
	}
}

func buildFrontPage(rd renderer.Renderer, outputDir string) error {
	frontPage, err := os.Create(fp.Join(outputDir, "index.html"))
	if err != nil {
		return fmt.Errorf("failed to create index file: %v", err)
	}

	err = rd.RenderFrontPage(frontPage)
	if err != nil {
		return fmt.Errorf("render failed: %v", err)
	}

	return nil
}

func buildList(rd renderer.Renderer, outputDir string, listType renderer.ListType, groupName string) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	for i := 0; ; i++ {
		fileName := "index.html"
		if i > 0 {
			fileName = fmt.Sprintf("%d.html", i)
		}
		fileName = fp.Join(outputDir, fileName)

		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", fileName, err)
		}

		nPosts, err := rd.RenderList(listType, groupName, i, f)
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to build list of posts: %v", err)
		}

		if nPosts == -1 {
			os.Remove(fileName)
			break
		}
	}

	return nil
}

func buildPages(rd renderer.Renderer, outputDir string, pages []model.Page) error {
	for _, page := range pages {
		page.Path = strings.TrimPrefix(page.Path, "/")

		dstDir := fp.Join(outputDir, page.Path)
		err := copyDir(page.Path, dstDir, "_index.md")
		if err != nil {
			return fmt.Errorf("failed to copy files for %s: %v", page.Path, err)
		}

		dstIndexPath := fp.Join(dstDir, "index.html")
		f, err := os.Create(dstIndexPath)
		if err != nil {
			return fmt.Errorf("failed to create index file for %s: %v", page.Path, err)
		}

		err = rd.RenderPage(page, f)
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to build %s: %v", page.Path, err)
		}
	}

	return nil
}

func buildPosts(rd renderer.Renderer, outputDir string, posts []model.Post) error {
	for i, post := range posts {
		post.Path = strings.TrimPrefix(post.Path, "/")

		dstDir := fp.Join(outputDir, post.Path)
		err := copyDir(post.Path, dstDir, "_index.md")
		if err != nil {
			return fmt.Errorf("failed to copy files for %s: %v", post.Path, err)
		}

		dstIndexPath := fp.Join(dstDir, "index.html")
		f, err := os.Create(dstIndexPath)
		if err != nil {
			return fmt.Errorf("failed to create index file for %s: %v", post.Path, err)
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
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to build %s: %v", post.Path, err)
		}
	}

	return nil
}
