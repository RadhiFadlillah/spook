package cmd

import (
	"fmt"
	"net/http"
	fp "path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RadhiFadlillah/spook/model"
	"github.com/RadhiFadlillah/spook/parser"
	"github.com/RadhiFadlillah/spook/renderer"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmdServer = &cobra.Command{
	Use:   "server",
	Short: "Run a webserver that serves the site",
	Long: "Run a simple and performant web server which serves the site. " +
		"Server will avoid writing the rendered and served content to disk, preferring to store it in memory. " +
		"If --port flag is not used, it will use port 8080 by default.",
	Aliases: []string{"serve"},
	Args:    cobra.NoArgs,
	Run:     cmdServerHandler,
}

func init() {
	cmdServer.Flags().IntP("port", "p", 8080, "Port that used by webserver")
}

func cmdServerHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	port, _ := cmd.Flags().GetInt("port")

	// Make sure valid config file exists in current working dir
	_, err := openConfigFile(true)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Create router
	router := httprouter.New()
	hdl := serverHandler{}

	router.GET("/js/*filepath", hdl.serveThemeFiles)
	router.GET("/res/*filepath", hdl.serveThemeFiles)
	router.GET("/css/*filepath", hdl.serveThemeFiles)
	router.GET("/static/*filepath", hdl.serveStaticFiles)

	router.GET("/", hdl.serveFrontPage)
	router.GET("/posts", hdl.serveListPage)
	router.GET("/posts/:n", hdl.serveListPage)
	router.GET("/category/:name", hdl.serveListPage)
	router.GET("/category/:name/:n", hdl.serveListPage)
	router.GET("/tag/:name", hdl.serveListPage)
	router.GET("/tag/:name/:n", hdl.serveListPage)
	router.GET("/page/:name", hdl.servePage)
	router.GET("/post/:name", hdl.servePost)

	// Route for panic
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Create server
	url := fmt.Sprintf(":%d", port)
	svr := &http.Server{
		Addr:         url,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	// Serve app
	logrus.Infoln("Serve spook in", url)
	logrus.Fatalln(svr.ListenAndServe())
}

type serverHandler struct{}

func (hdl *serverHandler) serveThemeFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	config, err := openConfigFile(true)
	checkError(err)

	filepath := fp.Join("theme", config.Theme, r.URL.Path)
	http.ServeFile(w, r, filepath)
}

func (hdl *serverHandler) serveStaticFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := fp.Join("static", ps.ByName("filepath"))
	http.ServeFile(w, r, filepath)
}

func (hdl *serverHandler) serveFrontPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Open config file
	config, err := openConfigFile(true)
	checkError(err)
	config.Theme = fp.Join("theme", config.Theme)

	// Parse all posts and pages
	posts, categories, tags, err := parser.ParsePosts(config)
	checkError(err)

	pages, err := parser.ParsePages(config)
	checkError(err)

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      posts,
		Tags:       tags,
		Categories: categories,
	}

	err = rd.RenderFrontPage(w)
	checkError(err)
}

func (hdl *serverHandler) serveListPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get parameter
	groupName := ps.ByName("name")
	strPageNumber := ps.ByName("n")
	pageNumber, _ := strconv.Atoi(strPageNumber)
	if pageNumber < 1 {
		pageNumber = 1
	}

	// Open config file
	config, err := openConfigFile(true)
	checkError(err)
	config.Theme = fp.Join("theme", config.Theme)

	// Parse all posts and pages
	posts, categories, tags, err := parser.ParsePosts(config)
	checkError(err)

	pages, err := parser.ParsePages(config)
	checkError(err)

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      posts,
		Tags:       tags,
		Categories: categories,
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

func (hdl *serverHandler) servePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Open config file
	config, err := openConfigFile(true)
	checkError(err)
	config.Theme = fp.Join("theme", config.Theme)

	// Parse all posts and pages
	posts, categories, tags, err := parser.ParsePosts(config)
	checkError(err)

	pages, err := parser.ParsePages(config)
	checkError(err)

	// Find the wanted page
	page := model.Page{}
	for i := 0; i < len(pages); i++ {
		if pages[i].Path == r.URL.Path {
			page = pages[i]
			page.Path = strings.TrimPrefix(page.Path, "/")
			break
		}
	}

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      posts,
		Tags:       tags,
		Categories: categories,
	}

	err = rd.RenderPage(page, w)
	checkError(err)
}

func (hdl *serverHandler) servePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Open config file
	config, err := openConfigFile(true)
	checkError(err)
	config.Theme = fp.Join("theme", config.Theme)

	// Parse all posts and pages
	posts, categories, tags, err := parser.ParsePosts(config)
	checkError(err)

	pages, err := parser.ParsePages(config)
	checkError(err)

	// Find the wanted post
	postIndex := -1
	for i := 0; i < len(posts); i++ {
		if posts[i].Path == r.URL.Path {
			postIndex = i
			break
		}
	}

	if postIndex == -1 {
		panic(fmt.Errorf("Post is not found"))
	}

	currentPost := posts[postIndex]
	currentPost.Path = strings.Trim(currentPost.Path, "/")

	newerPost := model.Post{}
	olderPost := model.Post{}

	if postIndex > 0 {
		newerPost = posts[postIndex-1]
	}

	if postIndex < len(posts)-1 {
		olderPost = posts[postIndex+1]
	}

	// Render and serve HTML
	rd := renderer.Renderer{
		Config:     config,
		Pages:      pages,
		Posts:      posts,
		Tags:       tags,
		Categories: categories,
	}

	err = rd.RenderPost(currentPost, olderPost, newerPost, w)
	checkError(err)
}
