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
	"github.com/Maxi-Mega/s3-image-server-v2/internal/metrics"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type Server struct {
	uiCfg          config.UI
	gatherer       *metrics.Metrics
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

func NewServer(cfg config.Config, cache types.Cache, frontendFS embed.FS, gatherer *metrics.Metrics, prod bool, version string) (*Server, error) {
	mode := "debug"
	if prod {
		mode = "production"
	}

	logger.Debug("Initializing web server in ", mode, " mode ...")

	faviconBase64, err := getBase64Content(cfg.UI.FaviconPngBase64Path)
	if err != nil {
		return nil, fmt.Errorf("favicon: %w", err)
	}

	logoBase64, err := getBase64Content(cfg.UI.LogoPngBase64Path)
	if err != nil {
		return nil, fmt.Errorf("logo: %w", err)
	}

	subFrontendFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		return nil, fmt.Errorf("failed to make sub-fs: %w", err)
	}

	subAssetsFS, err := fs.Sub(subFrontendFS, "assets")
	if err != nil {
		return nil, fmt.Errorf("failed to make assets sub-fs: %w", err)
	}

	graphResolver := &graph.Resolver{
		Cache: cache,
	}

	staticInfo := StaticInfo{
		SoftwareVersion:        version,
		WindowTitle:            cfg.UI.WindowTitle,
		ApplicationTitle:       cfg.UI.ApplicationTitle,
		FaviconBase64:          faviconBase64,
		LogoBase64:             logoBase64,
		ScaleInitialPercentage: int(cfg.UI.ScaleInitialPercentage),
		MaxImagesDisplayCount:  int(cfg.UI.MaxImagesDisplayCount),
		TileServerURL:          cfg.UI.Map.TileServerURL,
	}

	err = mapstructure.Decode(cfg.Products.ImageGroups, &staticInfo.ImageGroups) //nolint:musttag
	if err != nil {
		return nil, fmt.Errorf("failed to convert image groups to static info: %w", err)
	}

	srv := &Server{
		uiCfg:          cfg.UI,
		gatherer:       gatherer,
		addr:           fmt.Sprintf(":%d", cfg.UI.WebServerPort),
		cache:          cache,
		frontendFS:     frontendFS,
		subFrontendFS:  subFrontendFS,
		assetsFS:       subAssetsFS,
		graphqlHandler: handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver})),
		staticInfo:     staticInfo,
		wsHub:          newWSHub(),
	}

	return srv, srv.defineRoutes(prod)
}

func (srv *Server) Start(ctx context.Context, eventsChan chan types.OutEvent) error {
	srv.wsHub.goRun(ctx, eventsChan)

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

	logger.Info("Starting web server on http://localhost" + srv.addr + srv.uiCfg.BaseURL)

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err //nolint:wrapcheck
}
