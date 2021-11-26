package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"vervet-underground/config"

	gorillaMux "github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"vervet-underground/lib"
)

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx := context.Background()
	router := gorillaMux.NewRouter()
	var cfg *lib.ServerConfig
	var err error
	if cfg, err = config.Load("config.default.json"); err != nil {
		logError(err)
		panic("unable to load config")
	}

	log.Info().Msgf("services: %s", cfg.Services)
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

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Info().Msg(fmt.Sprintf("I'm starting my server on %s:8080", cfg.Host))
		if err := srv.ListenAndServe(); err != nil {
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
	log.Info().Str("message", "shutting down cleanly")
	os.Exit(0)
}

func logError(err error) {
	log.
		Error().
		Stack().
		Err(err).
		Str("cause", fmt.Sprintf("%+v", errors.Cause(err))).
		Msg("UnhandledException")
}
