package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	config "donetick.com/core/config"
	"github.com/gin-gonic/gin"
)

var embeddedFiles embed.FS

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
func staticMiddlewareNoRoute(root string) gin.HandlerFunc {
	fileServer := http.FileServer(getFileSystem(root))

	return func(c *gin.Context) {
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
