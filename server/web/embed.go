package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/daylamtayari/cierge/server/internal/util"
)

//go:embed dist
var assets embed.FS

// API path prefixes that must never fall back to the SPA — a typo'd API path
// should return a JSON 404, not a page of HTML.
var apiPathPrefixes = []string{"/api/", "/auth/", "/internal/"}

// ServeSPA returns a gin handler that serves the embedded React SPA.
// Client-side routes fall back to index.html. Unknown asset files (anything
// with an extension) and unknown API paths return 404 instead.
func ServeSPA() gin.HandlerFunc {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path

		for _, prefix := range apiPathPrefixes {
			if strings.HasPrefix(reqPath, prefix) {
				util.RespondNotFound(c, "")
				return
			}
		}

		p := strings.TrimPrefix(reqPath, "/")
		if p == "" {
			p = "."
		}
		if _, err := sub.Open(p); err != nil {
			if path.Ext(p) != "" {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Request.URL.Path = "/"
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
