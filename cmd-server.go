package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	fp "path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run a webserver that serves the site",
		Long: "Run a simple and performant web server which serves the site. " +
			"Server will avoid writing the rendered and served content to disk, preferring to store it in memory. " +
			"If --port flag is not used, it will use port 8080 by default.",
		Aliases: []string{"serve"},
		Args:    cobra.NoArgs,
		Run:     serverCmdHandler,
	}

	funcsMap = template.FuncMap{
		"formatTime":    formatTime,
		"limitSentence": limitSentence,
	}
)

type serverHandler struct {
	Config
}

func init() {
	serverCmd.Flags().IntP("port", "p", 8080, "Port that used by webserver")
}

func serverCmdHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	port, _ := cmd.Flags().GetInt("port")

	// Make sure valid config file exists in current working dir
	config := Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	if config.BaseURL == "" {
		cError.Println("Error: No base URL set in configuration file")
		return
	}

	if config.Theme == "" {
		cError.Println("Error: No theme specified in configuration file")
		return
	}

	if _, err = os.Stat(fp.Join("theme", config.Theme)); os.IsNotExist(err) {
		cError.Println("Error: The specified theme is not exists")
		return
	}

	// Create router
	router := httprouter.New()
	hdl := serverHandler{Config: config}

	router.GET("/js/*filepath", hdl.serveThemeFiles)
	router.GET("/res/*filepath", hdl.serveThemeFiles)
	router.GET("/css/*filepath", hdl.serveThemeFiles)
	router.GET("/static/*filepath", hdl.serveStaticFiles)

	router.GET("/", hdl.serveIndexPage)

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

func (hdl *serverHandler) serveThemeFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := fp.Join("theme", hdl.Theme, r.URL.Path)
	http.ServeFile(w, r, filepath)
}

func (hdl *serverHandler) serveStaticFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := fp.Join("static", ps.ByName("filepath"))
	http.ServeFile(w, r, filepath)
}

func (hdl *serverHandler) serveIndexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse all posts and page
	parsed, err := ParseAllPosts(hdl.Config, "post")
	if err != nil {
		panic(err)
	}

	// Create layout
	layout := NewLayoutIndex(1, hdl.Config, parsed)

	// Open and execute template
	themeDir := fp.Join("theme", hdl.Theme)
	tplIndexPath := fp.Join(themeDir, "index.html")
	tplPaths := append(getBaseTemplate(themeDir), tplIndexPath)

	tpl, err := template.New("").Funcs(funcsMap).ParseFiles(tplPaths...)
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "index.html", &layout)
	if err != nil {
		panic(err)
	}
}
