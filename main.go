package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/server"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web"
)

var (
	version   = "dev"
	buildTime = "now"                   //nolint: gochecknoglobals
	prod      = "false"                 //nolint: gochecknoglobals
	isProd, _ = strconv.ParseBool(prod) //nolint: gochecknoglobals
)

func main() {
	configPath := flag.String("c", "", "config file path")
	justPrintVersion := flag.Bool("v", false, "just print software version")
	justPrintDocs := flag.Bool("d", false, "just print documentation")

	flag.Usage = func() {
		fmt.Println("S3 Image Server", version, "- usage") //nolint:forbidigo
		flag.PrintDefaults()
		fmt.Print("\n- - Sample configuration - -\n\n") //nolint:forbidigo

		if err := printSampleConfig(os.Stdout); err != nil {
			log.Fatalln("Failed to print sample config:", err)
		}
	}

	flag.Parse()

	switch {
	case len(flag.Args()) != 0:
		log.Println("Invalid usage")
		flag.Usage()
		os.Exit(1)
	case *justPrintVersion:
		build := "development"

		if isProd {
			build = "production"
		}

		fmt.Printf("S3 Image Server %s, %s build / built at %s\n", version, build, buildTime) //nolint:forbidigo
		os.Exit(0)
	case *justPrintDocs:
		fmt.Print("S3 Image Server", version, "- documentation\n\n") //nolint:forbidigo

		if err := printDocumentation(os.Stdout); err != nil {
			log.Fatalln("Failed to print documentation:", err)
		}

		os.Exit(0)
	case *configPath == "":
		log.Fatalln("No configuration file path provided. Use -c <path> to specify one.")
	}

	cfg, warnings, err := config.Load(*configPath)
	if err != nil {
		log.Fatalln("Can't load configuration:", err)
	}

	err = logger.Init(cfg.Log.LogLevel, cfg.Log.ColorLogs, cfg.Log.JSONLogFormat, cfg.Log.JSONLogFields)
	if err != nil {
		log.Fatalln("Can't initialize logger:", err)
	}

	if len(warnings) > 0 {
		logger.Warnf("Configuration warnings:\n- %s", strings.Join(warnings, "\n- "))
	}

	start(cfg)
}

func start(cfg config.Config) {
	logger.Info("Starting S3 Image Server ", version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	metricGatherer := observability.New(cfg.Monitoring)

	srv, err := server.New(cfg, metricGatherer)
	if err != nil {
		logger.Fatal("Can't initialize server: ", err)
	}

	cache, outEvents, err := srv.Start(ctx)
	if err != nil {
		logger.Fatal("Can't start server: ", err)
	}

	webSrv, err := web.NewServer(cfg, cache, frontend, metricGatherer, isProd, version)
	if err != nil {
		logger.Fatal("Can't initialize web server: ", err)
	}

	err = webSrv.Start(ctx, outEvents)
	if err != nil {
		logger.Fatal("Can't start web server: ", err)
	}

	logger.Info("Shutting down the server.")
}
