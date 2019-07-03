package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
)

var (
	httpAddr  = ":8080"
	httpGrace = time.Second * 10
)

func main() {
	log.SetHandler(logfmt.Default)
	flag.StringVar(&httpAddr, "http", httpAddr, "HTTP listening address")
	flag.DurationVar(&httpGrace, "grace", httpGrace, "HTTP shutdown grace period for existing connections")
	flag.Parse()

	err := run()
	if err != nil {
		log.WithError(err).Fatal("fatal error")
	}
	log.Info("exit")
}

func run() error {
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: http.NotFoundHandler(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.WithField("timeout", httpGrace).Info("interrupt signal received. shutting down HTTP server with timeout for existing connections.")
		ctx, cancel := context.WithTimeout(context.Background(), httpGrace)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.WithError(err).Error("HTTP server Shutdown")
		}
		close(idleConnsClosed)
	}()

	log.WithField("http_addr", httpAddr).Info("HTTP server starting to listen")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener
		return err
	}
	<-idleConnsClosed
	return nil
}
