package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/server"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web"
)

var (
	version   = "dev"
	prod      = "false"                 //nolint: gochecknoglobals
	isProd, _ = strconv.ParseBool(prod) //nolint: gochecknoglobals
)

func main() {
	configPath := flag.String("c", "", "config file path")
	justPrintVersion := flag.Bool("v", false, "just print software version")

	flag.Usage = func() {
		fmt.Println("S3 Image Server V2", version, "- usage") //nolint:forbidigo
		flag.PrintDefaults()
		fmt.Println("\n- - Sample configuration - -") //nolint:forbidigo

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
		fmt.Println("S3 Image Server V2", version) //nolint:forbidigo
		os.Exit(0)
	case *configPath == "":
		log.Fatalln("No configuration file path provided. Use -c <path> to specify one.")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalln("Can't load configuration:", err)
	}

	err = logger.Init(cfg.Log)
	if err != nil {
		log.Fatalln("Can't initialize logger:", err)
	}

	start(cfg)
}

func start(cfg config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Can't initialize server: ", err)
	}

	cache, outEvents, err := srv.Start(ctx)
	if err != nil {
		logger.Fatal("Can't start server: ", err)
	}

	webSrv, err := web.NewServer(cfg.UI, cache, frontend, isProd)
	if err != nil {
		logger.Fatal("Can't initialize web server: ", err)
	}

	err = webSrv.Start(ctx, outEvents)
	if err != nil {
		logger.Fatal("Can't start web server: ", err)
	}

	logger.Info("Shutting down the server.")
}
