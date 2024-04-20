package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"

	"github.com/gin-gonic/gin"
)

var (
	errNoCacheKey      = errors.New("no cache key provided")
	errInvalidCacheKey = errors.New("invalid cache key")
	errUnexpected      = errors.New("an unexpected error occurred - check the server logs for more information")
)

type Error struct {
	Err error `json:"error"`
}

type StaticInfo struct {
	WindowTitle            string `json:"windowTitle"`
	ApplicationTitle       string `json:"applicationTitle"`
	FaviconBase64          string `json:"faviconBase64"`
	LogoBase64             string `json:"logoBase64"`
	ScaleInitialPercentage int    `json:"scaleInitialPercentage"`
	MaxImagesDisplayCount  int    `json:"maxImagesDisplayCount"`
	TileServerURL          string `json:"tileServerURL"`
	ImageGroups            []struct {
		GroupName string `json:"name"`
		Bucket    string `json:"bucket"`
		Types     []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"types"`
	} `json:"imageGroups"`
}

func (srv *Server) infoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, srv.staticInfo)
}

func (srv *Server) cacheHandler(c *gin.Context) {
	cacheKey := strings.Trim(c.Param("cache_key"), "/")
	if cacheKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{errNoCacheKey})

		return
	}

	if strings.Contains(cacheKey, "..") {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{errInvalidCacheKey})

		return
	}

	objectData, err := srv.cache.GetCachedObject(cacheKey)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			c.AbortWithStatusJSON(http.StatusNotFound, Error{fmt.Errorf("cache key %q not found", cacheKey)})
		} else {
			logger.Infof("Unexpected error while serving cache key %q: %v", cacheKey, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, Error{errUnexpected})
		}

		return
	}

	c.Data(http.StatusOK, "application/octet-stream", objectData)
}
