package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"vervet-underground/config"
	"vervet-underground/internal/scraper"
	"vervet-underground/internal/storage/mem"
)

func main() {
	var wait time.Duration
	var configJson string
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.StringVar(&configJson, "config-file", "config.default.json", "the configuration file holding target services and the host address to run server on")

	flag.Parse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	router := mux.NewRouter()
	var cfg *config.ServerConfig
	var err error
	if cfg, err = config.Load(configJson); err != nil {
		logError(err)
		panic("unable to load config")
	}
	log.Info().Msgf("services: %s", cfg.Services)

	// initialize Scraper
	ticker := time.NewTicker(time.Minute * 5)
	st := mem.New()
	sc, err := scraper.New(cfg, st, scraper.HTTPClient(&http.Client{Timeout: wait}))
	if err != nil {
		logError(err)
		panic("unable to load storage")
	}
	// initialize
	err = runScrape(wait, sc)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initialization scraping of service")
	}

	versionHandlers(router, sc)
	router.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(fmt.Sprintf(`{"msg": "success", "services": %s}`, cfg.Services)))
		if err != nil {
			http.Error(w, "Failure to write response", http.StatusInternalServerError)
			return
		}
	})

	srv := &http.Server{
		Addr: fmt.Sprintf("%s:8080", cfg.Host),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	quit := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := runScrape(wait, sc)
				if err != nil {
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func runScrape(wait time.Duration, sc *scraper.Scraper) error {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	scraperErr := sc.Run(ctx)
	if scraperErr != nil {
		return scraperErr
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

func versionHandlers(router *mux.Router, sc *scraper.Scraper) {
	router.
		Path("/openapi").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			versionSlice, err := json.Marshal(sc.Versions())
			if err != nil {
				logError(err)
				http.Error(w, "Failure to process request", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(versionSlice)
			if err != nil {
				logError(err)
				http.Error(w, "Failure to write response", http.StatusInternalServerError)
				return
			}
		})

	router.
		Path("/openapi/{version}").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := mux.Vars(r)["version"]
			bytes, err := sc.Version(version)
			if err != nil {
				logError(err)
				http.Error(w, "Failure to process request", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(bytes)
			if err != nil {
				logError(err)
				http.Error(w, "Failure to write response", http.StatusInternalServerError)
				return
			}
		})
}
