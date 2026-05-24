package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var assets embed.FS

// ServeSPA returns a gin handler that serves the embedded React SPA.
// All non-asset requests fall back to index.html for client-side routing.
func ServeSPA() gin.HandlerFunc {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return func(c *gin.Context) {
		_, err := sub.Open(c.Request.URL.Path)
		if err != nil {
			c.Request.URL.Path = "/"
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
