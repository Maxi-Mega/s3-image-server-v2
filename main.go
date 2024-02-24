package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
)

var version string

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
		log.Println("S3 Image Server V2", version)
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
}
