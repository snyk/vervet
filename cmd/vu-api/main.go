package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/snyk/vervet/v6/config"
	"github.com/snyk/vervet/v6/internal/handler"
	"github.com/snyk/vervet/v6/internal/storage"
	"github.com/snyk/vervet/v6/internal/storage/disk"
	"github.com/snyk/vervet/v6/internal/storage/gcs"
	"github.com/snyk/vervet/v6/internal/storage/s3"
)

func main() {
	var wait time.Duration
	var configJson string
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15,
		"the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.StringVar(&configJson, "config-file", "config.default.json",
		"the configuration file holding target services and the host address to run server on")

	flag.Parse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var cfg *config.ServerConfig
	var err error
	if cfg, err = config.LoadServerConfig(configJson); err != nil {
		log.Fatal().Err(err).Msg("unable to load config")
	}

	ctx := context.Background()
	st, err := initializeStorage(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to initialize storage client")
	}

	h := handler.New(cfg, st, handler.UseDefaultMiddleware)

	srv := &http.Server{
		Addr: fmt.Sprintf("%s:8080", cfg.Host),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h,
	}

	grp, grpCtx := errgroup.WithContext(ctx)
	grp.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		log.Info().Msg(fmt.Sprintf("I'm starting my server on %s:8080", cfg.Host))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start server: %w", err)
		}
		return nil
	})

	grp.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		interceptSignals(grpCtx)
		time.Sleep(wait)
		log.Info().Float64("timeoutSeconds", wait.Seconds()).Msg("server stopping")

		return shutdown(ctx, srv, wait)
	})

	if err := grp.Wait(); err != nil {
		log.Fatal().Err(err).Msg("server unexpectedly stopped")
	}

	log.Info().Msg("shutting down cleanly")
}

func initializeStorage(ctx context.Context, cfg *config.ServerConfig) (storage.ReadOnlyStorage, error) {
	switch cfg.Storage.Type {
	case config.StorageTypeDisk:
		return disk.New(cfg.Storage.Disk.Path), nil
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
		})
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
		})
	}
	return nil, fmt.Errorf("unknown storage backend: %s", cfg.Storage.Type)
}

func interceptSignals(ctx context.Context) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	select {
	case <-ctx.Done():
	case sig := <-sigc:
		log.Info().Str("signal", sig.String()).Msg("intercepted signal")
	}
}

func shutdown(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}

	return nil
}
