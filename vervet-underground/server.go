package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"vervet-underground/config"
	"vervet-underground/internal/handler"
	"vervet-underground/internal/scraper"
	"vervet-underground/internal/storage"
	"vervet-underground/internal/storage/disk"
	"vervet-underground/internal/storage/gcs"
	"vervet-underground/internal/storage/s3"
)

func main() {
	var wait time.Duration
	var scrapeInterval time.Duration
	var configJson string
	var overlayFile string
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15,
		"the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.DurationVar(&scrapeInterval, "scrape-interval", time.Minute,
		"the frequency at which scraping occurs  - e.g. 15s, 1m, 1h")
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
		logError(err)
		log.Fatal().Msg("unable to load config")
	}
	log.Info().Msgf("services: %s", cfg.Services)

	var overlayContents []byte
	if overlayFile != "" {
		overlayContents, err = os.ReadFile(overlayFile)
		if err != nil {
			logError(err)
			log.Fatal().Msgf("unable to load overlay file %q", overlayFile)
		}
	}

	// initialize Scraper
	ticker := time.NewTicker(scrapeInterval)
	ctx := context.Background()
	st, err := initializeStorage(ctx, cfg, overlayContents)
	if err != nil {
		logError(err)
		log.Fatal().Msg("unable to initialize storage client")
	}

	sc, err := scraper.New(cfg, st, scraper.HTTPClient(&http.Client{
		Timeout:   wait,
		Transport: scraper.DurationTransport(http.DefaultTransport),
	}))
	if err != nil {
		logError(err)
		log.Fatal().Msg("unable to load storage")
	}
	// initialize
	err = runScrape(ctx, sc)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initialization scraping of service")
	}

	h := handler.New(cfg, sc, handler.UseDefaultMiddleware)

	srv := &http.Server{
		Addr: fmt.Sprintf("%s:8080", cfg.Host),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h,
	}

	quit := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := runScrape(ctx, sc); err != nil {
					logError(err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Info().Msg(fmt.Sprintf("I'm starting my server on %s:8080", cfg.Host))
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logError(err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c

	// closes the scraper go routine
	quit <- struct{}{}
	close(quit)
	log.Info().Msg("scraper successfully spun down")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err = srv.Shutdown(ctx)
	if err != nil {
		logError(err)
	}

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info().Msg("shutting down cleanly")
	os.Exit(0)
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

func logError(err error) {
	log.
		Error().
		Stack().
		Err(err).
		Str("cause", fmt.Sprintf("%+v", errors.Cause(err))).
		Msg("UnhandledException")
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
