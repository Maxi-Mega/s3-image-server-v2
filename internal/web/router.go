package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func (srv *Server) defineRoutes(prod bool) error {
	var e *gin.Engine

	if prod {
		gin.SetMode(gin.ReleaseMode)

		e = gin.New()
		e.Use(gin.Recovery())
	} else {
		gin.SetMode(gin.DebugMode)

		e = gin.Default() // Default includes logger & recovery middlewares
	}

	// TODO: prometheus middleware & endpoint

	r := e.Group(srv.uiCfg.BaseURL)

	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Frontend
	r.GET("/", srv.frontHandler)
	r.StaticFS("/assets", http.FS(srv.assetsFS))

	// API
	api := r.Group("/api")
	api.GET("/info", srv.infoHandler)
	api.GET("/cache/:cacheKey", srv.cacheHandler)
	api.GET("/ws", srv.wsHub.serveWs)
	api.POST("/graphql", handlerAdapter(srv.graphqlHandler))

	if !prod {
		playgroundHandler, err := srv.makePlaygroundHandler()
		if err != nil {
			return err
		}

		api.GET("/playground", playgroundHandler)
	}

	// TODO: stats endpoint

	srv.router = e

	return nil
}

func (srv *Server) withBasePath(route string) string {
	result, err := url.JoinPath(srv.uiCfg.BaseURL, route)
	if err != nil {
		logger.Errorf("Failed to join %q with base path: %v", route, err)

		return route
	}

	return result
}

func (srv *Server) makePlaygroundHandler() (gin.HandlerFunc, error) {
	queryURL, err := url.JoinPath(srv.uiCfg.BaseURL, "/api/graphql")
	if err != nil {
		return nil, fmt.Errorf("failed to join /api/query with base path: %w", err)
	}

	return handlerAdapter(playground.Handler("GraphQL playground", queryURL)), nil
}
