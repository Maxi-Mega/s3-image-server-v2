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
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
)

type Server struct {
	uiCfg          config.UI
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

func NewServer(uiCfg config.UI, cache types.Cache, frontendFS embed.FS, prod bool) (*Server, error) {
	mode := "debug"
	if prod {
		mode = "production"
	}

	logger.Debug("Initializing web server in ", mode, " mode ...")

	logoBase64, err := getLogoBase64(uiCfg.LogoBase64Path)
	if err != nil {
		return nil, err
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

	srv := &Server{
		uiCfg:          uiCfg,
		addr:           fmt.Sprintf(":%d", uiCfg.WebServerPort),
		cache:          cache,
		frontendFS:     frontendFS,
		subFrontendFS:  subFrontendFS,
		assetsFS:       subAssetsFS,
		graphqlHandler: handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graphResolver})),
		staticInfo: StaticInfo{
			WindowTitle:            uiCfg.WindowTitle,
			ApplicationTitle:       uiCfg.ApplicationTitle,
			LogoBase64:             logoBase64,
			ScaleInitialPercentage: int(uiCfg.ScaleInitialPercentage),
			MaxImagesDisplayCount:  int(uiCfg.MaxImagesDisplayCount),
			TileServerURL:          uiCfg.Map.TileServerURL,
		},
		wsHub: newWSHub(),
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

	logger.Info("Starting web server on http://localhost" + srv.addr)

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err //nolint:wrapcheck
}
