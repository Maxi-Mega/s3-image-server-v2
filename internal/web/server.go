package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
)

type Server struct {
	uiCfg          config.UI
	gatherer       *observability.Metrics
	addr           string
	cache          types.Cache
	frontendFS     embed.FS
	subFrontendFS  fs.FS
	assetsFS       fs.FS
	graphqlHandler *handler.Server
	staticInfo     StaticInfo
	router         *gin.Engine
	wsHub          *wsHub
}

func NewServer(cfg config.Config, cache types.Cache, frontendFS embed.FS, gatherer *observability.Metrics, prod bool, version string) (*Server, error) {
	mode := "debug"
	if prod {
		mode = "production"
	}

	logger.Debug("Initializing web server in ", mode, " mode ...")

	subFrontendFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		return nil, fmt.Errorf("failed to make sub-fs: %w", err)
	}

	subAssetsFS, err := fs.Sub(subFrontendFS, "assets")
	if err != nil {
		return nil, fmt.Errorf("failed to make assets sub-fs: %w", err)
	}

	graphResolver := &graph.Resolver{
		Config: cfg,
		Cache:  cache,
	}

	staticInfo := StaticInfo{
		SoftwareVersion:        version,
		WindowTitle:            cfg.UI.WindowTitle,
		ApplicationTitle:       cfg.UI.ApplicationTitle,
		FaviconBase64:          cfg.UI.FaviconPngBase64,
		LogoBase64:             cfg.UI.LogoPngBase64,
		ScaleInitialPercentage: int(cfg.UI.ScaleInitialPercentage), //nolint: gosec
		MaxImagesDisplayCount:  int(cfg.UI.MaxImagesDisplayCount),  //nolint: gosec
		PMTilesURL:             cfg.UI.Map.PMTilesURL,
		PMTilesStyleURL:        cfg.UI.Map.PMTilesStyleURL,
	}

	err = mapstructure.Decode(cfg.Products.ImageGroups, &staticInfo.ImageGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image groups to static info: %w", err)
	}

	graphqlHandler := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver}))
	graphqlHandler.AddTransport(transport.Options{})
	graphqlHandler.AddTransport(transport.POST{})

	if !prod {
		graphqlHandler.Use(extension.Introspection{})
	}

	srv := &Server{
		uiCfg:          cfg.UI,
		gatherer:       gatherer,
		addr:           fmt.Sprintf(":%d", cfg.UI.WebServerPort),
		cache:          cache,
		frontendFS:     frontendFS,
		subFrontendFS:  subFrontendFS,
		assetsFS:       subAssetsFS,
		graphqlHandler: graphqlHandler,
		staticInfo:     staticInfo,
		wsHub:          newWSHub(),
	}

	return srv, srv.defineRoutes(prod)
}

func (srv *Server) Start(ctx context.Context, eventsChan chan types.OutEvent) error {
	srv.wsHub.goRun(ctx, eventsChan)

	eventsChan <- types.OutEvent{EventType: types.EventReset}

	httpServer := &http.Server{
		Addr:              srv.addr,
		Handler:           srv.router,
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_ = httpServer.Shutdown(shutdownCtx) //nolint:contextcheck
	}()

	logger.Infof("Starting web server on %s (base URL %q)", srv.addr, srv.uiCfg.BaseURL)

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err //nolint:wrapcheck
}
