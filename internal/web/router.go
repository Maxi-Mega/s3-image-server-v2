package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (srv *Server) defineRoutes(prod bool) error {
	var e *gin.Engine

	if prod {
		gin.SetMode(gin.ReleaseMode)

		e = gin.New()
		e.Use(gin.Recovery())
	} else {
		gin.SetMode(gin.DebugMode)

		e = gin.Default()     // Default includes logger & recovery middlewares
		e.Use(cors.Default()) // To avoid CORS issues when the frontend is started with its own server
	}

	r := e.Group(srv.uiCfg.BaseURL)

	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Frontend
	front := r.Use(metricsMiddleware(srv.gatherer, endpointFront))
	front.GET("/", srv.frontHandler)
	front.GET("/favicon.ico", srv.frontHandler)
	front.StaticFS("/assets", http.FS(srv.assetsFS))

	// API
	api := r.Group("/api").Use(metricsMiddleware(srv.gatherer, endpointAPI))
	api.GET("/info", srv.infoHandler)
	api.GET("/cache/*cache_key", srv.cacheHandler)
	api.GET("/ws", srv.wsHub.serveWs)
	api.POST("/graphql", gin.WrapH(srv.graphqlHandler))

	if !prod {
		api.GET("/dump-images", func(c *gin.Context) {
			c.JSON(200, srv.cache.DumpImages())
		})

		playgroundHandler, err := srv.makePlaygroundHandler()
		if err != nil {
			return err
		}

		api.GET("/playground", playgroundHandler)
	}

	promHandler := promhttp.Handler()
	r.GET("/metrics", gin.WrapH(promHandler))

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

	return gin.WrapH(playground.Handler("GraphQL playground", queryURL)), nil
}
