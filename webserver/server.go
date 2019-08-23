package webserver

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-spook/spook/model"
	"github.com/julienschmidt/httprouter"
)

// Start serves blog in specified port
func Start(rootDir string, config model.Config, port int) error {
	// Create handler
	hdl := handler{
		Config:  config,
		RootDir: rootDir,
	}

	// Create router
	router := httprouter.New()

	router.GET("/js/*filepath", hdl.serveThemeFiles)
	router.GET("/res/*filepath", hdl.serveThemeFiles)
	router.GET("/css/*filepath", hdl.serveThemeFiles)
	router.GET("/static/*filepath", hdl.serveStaticFiles)

	router.GET("/", hdl.serveFrontPage)
	router.GET("/posts", hdl.serveList)
	router.GET("/posts/:n", hdl.serveList)
	router.GET("/category/:name", hdl.serveList)
	router.GET("/category/:name/:n", hdl.serveList)
	router.GET("/tag/:name", hdl.serveList)
	router.GET("/tag/:name/:n", hdl.serveList)
	router.GET("/page/:name", hdl.addSuffixSlash)
	router.GET("/page/:name/*filepath", hdl.servePage)
	router.GET("/post/:name", hdl.addSuffixSlash)
	router.GET("/post/:name/*filepath", hdl.servePost)

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
	return svr.ListenAndServe()
}

func checkError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}
