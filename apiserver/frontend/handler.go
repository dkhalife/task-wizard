package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	config "dkhalife.com/tasks/core/config"
	"github.com/gin-gonic/gin"
)

//go:embed dist
var embeddedFiles embed.FS

var publicPaths = []string{"/login", "/privacy"}

type Handler struct {
	ServeFrontend bool
}

func NewHandler(config *config.Config) *Handler {
	return &Handler{
		ServeFrontend: config.Server.ServeFrontend,
	}
}

func Routes(router *gin.Engine, h *Handler) {
	if h.ServeFrontend {
		router.Use(staticMiddleware("dist"))
		router.NoRoute(staticMiddlewareNoRoute("dist"))
	}
}

func isPublicPath(path string) bool {
	for _, p := range publicPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

func staticMiddleware(root string) gin.HandlerFunc {
	fileServer := http.FileServer(getFileSystem(root))

	return func(c *gin.Context) {
		_, err := fs.Stat(embeddedFiles, "dist"+c.Request.URL.Path)
		if err != nil {
			c.Next()
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)

	}
}
const authCookieName = "tw_auth"

func staticMiddlewareNoRoute(root string) gin.HandlerFunc {
	fileServer := http.FileServer(getFileSystem(root))

	return func(c *gin.Context) {
		if !isPublicPath(c.Request.URL.Path) {
			if _, err := c.Cookie(authCookieName); err != nil {
				target := "/login?return_to=" + c.Request.URL.Path
				if q := c.Request.URL.RawQuery; q != "" {
					target += "&" + q
				}
				c.Redirect(http.StatusFound, target)
				return
			}
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func getFileSystem(path string) http.FileSystem {
	fs, err := fs.Sub(embeddedFiles, path)
	if err != nil {
		panic(err)
	}
	return http.FS(fs)
}
