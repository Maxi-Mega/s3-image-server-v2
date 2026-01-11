package web

import (
	"net/http"
	"os"
	"strings"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"

	"github.com/gin-gonic/gin"
)

func (srv *Server) frontHandler(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, srv.withBasePath("/api/")) {
		http.NotFound(c.Writer, c.Request)

		return
	}

	route := strings.TrimPrefix(c.Request.URL.Path, srv.uiCfg.BaseURL)
	switch route {
	case "favicon.ico", "/favicon.ico":
		srv.serveFrontendResource(c, "favicon.ico", "image/x-icon")
	case "", "/", "doc":
		srv.serveFrontendResource(c, "index.html", "text/html")
	default:
		http.NotFound(c.Writer, c.Request)
	}
}

func (srv *Server) serveFrontendResource(c *gin.Context, name, contentType string) {
	rawFile, err := srv.subFrontendFS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(c.Writer, c.Request)

			return
		}

		logger.Fatalf("Can't read frontend resource %q: %v", name, err)
	}

	stat, err := rawFile.Stat()
	if err != nil {
		logger.Fatalf("Can't stat frontend resource %q: %v", name, err)
	}

	c.DataFromReader(http.StatusOK, stat.Size(), contentType, rawFile, nil)
}
