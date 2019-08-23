package webserver

import (
	"fmt"
	"net/http"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/go-spook/spook/model"
	"github.com/go-spook/spook/parser"
	"github.com/go-spook/spook/renderer"
	"github.com/julienschmidt/httprouter"
)

// handler is handler for serving the web interface.
type handler struct {
	Config  model.Config
	RootDir string
}

func (hdl *handler) serveThemeFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := fp.Join("theme", hdl.Config.Theme, r.URL.Path)
	http.ServeFile(w, r, filepath)
}

func (hdl *handler) serveStaticFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := fp.Join("static", ps.ByName("filepath"))
	http.ServeFile(w, r, filepath)
}

func (hdl *handler) serveFrontPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse all posts and pages
	psr := parser.Parser{
		Config:  hdl.Config,
		RootDir: hdl.RootDir,
	}

	parsedPosts, err := psr.ParsePosts()
	checkError(err)

	pages, err := psr.ParsePages()
	checkError(err)

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     hdl.Config,
		Pages:      pages,
		Posts:      parsedPosts.Posts,
		Tags:       parsedPosts.Tags,
		Categories: parsedPosts.Categories,
		RootDir:    hdl.RootDir,
	}

	err = rd.RenderFrontPage(w)
	checkError(err)
}

func (hdl *handler) serveList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get parameter
	groupName := ps.ByName("name")
	strPageNumber := ps.ByName("n")
	pageNumber, _ := strconv.Atoi(strPageNumber)
	if pageNumber < 1 {
		pageNumber = 1
	}

	// Parse all posts and pages
	psr := parser.Parser{
		Config:  hdl.Config,
		RootDir: hdl.RootDir,
	}

	parsedPosts, err := psr.ParsePosts()
	checkError(err)

	pages, err := psr.ParsePages()
	checkError(err)

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     hdl.Config,
		Pages:      pages,
		Posts:      parsedPosts.Posts,
		Tags:       parsedPosts.Tags,
		Categories: parsedPosts.Categories,
		RootDir:    hdl.RootDir,
	}

	var listType renderer.ListType
	if strings.HasPrefix(r.URL.Path, "/category") {
		listType = renderer.CATEGORY
	} else if strings.HasPrefix(r.URL.Path, "/tag") {
		listType = renderer.TAG
	} else {
		listType = renderer.DEFAULT
		groupName = ""
	}

	_, err = rd.RenderList(listType, groupName, pageNumber, w)
	checkError(err)
}

func (hdl *handler) servePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pagePath := fp.Join("page", ps.ByName("name"))

	// Check if this is request for asset file of a page
	if filepath := ps.ByName("filepath"); filepath != "" && filepath != "/" {
		filepath = fp.Join(pagePath, filepath)
		http.ServeFile(w, r, filepath)
		return
	}

	// Parse all posts and pages
	psr := parser.Parser{
		Config:  hdl.Config,
		RootDir: hdl.RootDir,
	}

	parsedPosts, err := psr.ParsePosts()
	checkError(err)

	pages, err := psr.ParsePages()
	checkError(err)

	// Find the wanted page
	page := model.Page{}
	for i := 0; i < len(pages); i++ {
		if pages[i].Path == "/"+pagePath {
			page = pages[i]
			page.Path = strings.TrimPrefix(page.Path, "/")
			break
		}
	}

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     hdl.Config,
		Pages:      pages,
		Posts:      parsedPosts.Posts,
		Tags:       parsedPosts.Tags,
		Categories: parsedPosts.Categories,
		RootDir:    hdl.RootDir,
	}

	err = rd.RenderPage(page, w)
	checkError(err)
}

func (hdl *handler) servePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	postPath := fp.Join("post", ps.ByName("name"))

	// Check if this is request for asset file of a post
	if filepath := ps.ByName("filepath"); filepath != "" && filepath != "/" {
		filepath = fp.Join(postPath, filepath)
		http.ServeFile(w, r, filepath)
		return
	}

	// Parse all posts and pages
	psr := parser.Parser{
		Config:  hdl.Config,
		RootDir: hdl.RootDir,
	}

	parsedPosts, err := psr.ParsePosts()
	checkError(err)

	pages, err := psr.ParsePages()
	checkError(err)

	// Find the wanted post
	postIndex := -1
	for i := 0; i < len(parsedPosts.Posts); i++ {
		if parsedPosts.Posts[i].Path == "/"+postPath {
			postIndex = i
			break
		}
	}

	if postIndex == -1 {
		panic(fmt.Errorf("post is not found"))
	}

	currentPost := parsedPosts.Posts[postIndex]
	currentPost.Path = strings.Trim(currentPost.Path, "/")

	newerPost := model.Post{}
	olderPost := model.Post{}

	if postIndex > 0 {
		newerPost = parsedPosts.Posts[postIndex-1]
	}

	if postIndex < len(parsedPosts.Posts)-1 {
		olderPost = parsedPosts.Posts[postIndex+1]
	}

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     hdl.Config,
		Pages:      pages,
		Posts:      parsedPosts.Posts,
		Tags:       parsedPosts.Tags,
		Categories: parsedPosts.Categories,
		RootDir:    hdl.RootDir,
	}

	err = rd.RenderPost(currentPost, olderPost, newerPost, w)
	checkError(err)
}

func (hdl *handler) addSuffixSlash(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	newPath := fp.Clean(r.URL.Path) + "/"
	http.Redirect(w, r, newPath, 301)
}
