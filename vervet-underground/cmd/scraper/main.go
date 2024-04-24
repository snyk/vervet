package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"vervet-underground/config"
	"vervet-underground/internal/scraper"
	"vervet-underground/internal/storage"
	"vervet-underground/internal/storage/disk"
	"vervet-underground/internal/storage/gcs"
	"vervet-underground/internal/storage/s3"
)

func main() {
	var wait time.Duration
	var configJson string
	var overlayFile string
	flag.StringVar(&configJson, "config-file", "config.default.json",
		"the configuration file holding target services and the host address to run server on")
	flag.StringVar(&overlayFile, "overlay-file", "",
		"OpenAPI document fragment overlay applied to all collated output")

	flag.Parse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var cfg *config.ServerConfig
	var err error
	if cfg, err = config.Load(configJson); err != nil {
		log.Fatal().Err(err).Msg("unable to load config")
	}
	log.Info().Msgf("services: %s", cfg.Services)

	var overlayContents []byte
	if overlayFile != "" {
		overlayContents, err = os.ReadFile(overlayFile)
		if err != nil {
			log.Fatal().Err(err).Msgf("unable to load overlay file %q", overlayFile)
		}
	}

	ctx := context.Background()
	st, err := initializeStorage(ctx, cfg, overlayContents)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initialize storage client")
	}

	sc, err := scraper.New(cfg, st, scraper.HTTPClient(&http.Client{
		Timeout:   wait,
		Transport: scraper.DurationTransport(http.DefaultTransport),
	}))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load storage")
	}

	err = runScrape(ctx, sc)
	if err != nil {
		log.Fatal().Err(err).Msg("failed scraping of service")
	}
}

// runScrape runs scraping all services and can take
// a longer period of time than standard wait timeout.
// moves to cancel context once scraping and collation are complete.
func runScrape(ctx context.Context, sc *scraper.Scraper) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := sc.Run(ctxWithCancel); err != nil {
		return err
	}
	log.Info().Msgf("scraper successfully completed run at %s", time.Now().UTC().String())
	return nil
}

func initializeStorage(ctx context.Context, cfg *config.ServerConfig, overlayContents []byte) (storage.Storage, error) {
	collatorOpts := []storage.CollatorOption{storage.CollatorExcludePattern(cfg.Merging.ExcludePatterns)}
	if overlayContents != nil {
		collatorOpts = append(collatorOpts, storage.CollatorOverlay(string(overlayContents)))
	}
	newCollator := func() (*storage.Collator, error) {
		return storage.NewCollator(collatorOpts...)
	}
	switch cfg.Storage.Type {
	case config.StorageTypeDisk:
		return disk.New(cfg.Storage.Disk.Path, disk.NewCollator(newCollator)), nil
	case config.StorageTypeS3:
		return s3.New(ctx, &s3.Config{
			AwsRegion:      cfg.Storage.S3.Region,
			AwsEndpoint:    cfg.Storage.S3.Endpoint,
			IamRoleEnabled: cfg.Storage.IamRoleEnabled,
			BucketName:     cfg.Storage.BucketName,
			Credentials: s3.StaticKeyCredentials{
				AccessKey:  cfg.Storage.S3.AccessKey,
				SecretKey:  cfg.Storage.S3.SecretKey,
				SessionKey: cfg.Storage.S3.SessionKey,
			},
		}, s3.NewCollator(newCollator))
	case config.StorageTypeGCS:
		return gcs.New(ctx, &gcs.Config{
			GcsRegion:      cfg.Storage.GCS.Region,
			GcsEndpoint:    cfg.Storage.GCS.Endpoint,
			IamRoleEnabled: cfg.Storage.IamRoleEnabled,
			BucketName:     cfg.Storage.BucketName,
			Credentials: gcs.StaticKeyCredentials{
				ProjectId: cfg.Storage.GCS.ProjectId,
				Filename:  cfg.Storage.GCS.Filename,
			},
		}, gcs.NewCollator(newCollator))
	}
	return nil, fmt.Errorf("unknown storage backend: %s", cfg.Storage.Type)
}
