package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

const defaultCacheDirName = "s3_image_server"

var (
	errInvalidConfig          = errors.New("the config is invalid")
	errNoImageGroupsSpecified = errors.New("no image groups specified")
	errDuplicate              = errors.New("duplicate")
)

func Load(configPath string) (Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, err //nolint:wrapcheck
	}

	defer file.Close()

	var cfg = defaultConfig()

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	err = cfg.validate()
	if err != nil {
		return Config{}, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	err = cfg.process()
	if err != nil {
		return Config{}, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	errs := make([]error, 0)

	if len(cfg.Products.ImageGroups) == 0 {
		errs = append(errs, errNoImageGroupsSpecified)
	}

	imageGroupNames := make(map[string]bool)

	for _, grp := range cfg.Products.ImageGroups {
		if imageGroupNames[grp.GroupName] {
			errs = append(errs, fmt.Errorf("image group name %q is %w", grp.GroupName, errDuplicate))

			break
		}

		imageGroupNames[grp.GroupName] = true
		imageTypeNames := make(map[string]bool)

		for _, typ := range grp.Types {
			if imageTypeNames[typ.Name] {
				errs = append(errs, fmt.Errorf("image type name %q of group %q is %w", typ.Name, grp.GroupName, errDuplicate))

				break
			}

			imageTypeNames[typ.Name] = true
		}
	}

	return errors.Join(errs...)
}

func (cfg *Config) process() (err error) {
	cfg.Cache.CacheDir, err = filepath.Abs(cfg.Cache.CacheDir)
	if err != nil {
		return fmt.Errorf("could not resolve cache dir: %w", err)
	}

	cfg.Cache.CacheDir = filepath.Join(cfg.Cache.CacheDir, defaultCacheDirName)

	if cfg.UI.BaseURL == "" {
		cfg.UI.BaseURL = "/"
	}

	cfg.Products.AdditionalProductFilesRgx, err = regexp.Compile(cfg.Products.AdditionalProductFilesRegexp)
	if err != nil {
		return fmt.Errorf("can't parse products.additionalProductFilesRegexp: %w", err)
	}

	cfg.Products.TargetRelativeRgx, err = regexp.Compile(cfg.Products.TargetRelativeRegexp)
	if err != nil {
		return fmt.Errorf("can't parse products.targetRelativeRegexp: %w", err)
	}

	cfg.Products.FeaturesExtensionRgx, err = regexp.Compile(cfg.Products.FeaturesExtensionRegexp)
	if err != nil {
		return fmt.Errorf("can't parse products.featuresExtensionRegexp: %w", err)
	}

	for g, imgGroup := range cfg.Products.ImageGroups {
		for t, imgType := range imgGroup.Types {
			cfg.Products.ImageGroups[g].Types[t].ProductRgx, err = regexp.Compile(imgType.ProductRegexp)
			if err != nil {
				return fmt.Errorf("can't parse products.imageGroups[%q].types[%q].productRegexp: %w", imgGroup.GroupName, imgType.Name, err)
			}
		}
	}

	return nil
}
