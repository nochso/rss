package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	httpAddr  = ":8080"
	httpGrace = time.Second * 10
)

func main() {
	flag.StringVar(&httpAddr, "http", httpAddr, "HTTP listening address")
	flag.DurationVar(&httpGrace, "grace", httpGrace, "HTTP shutdown grace period for existing connections")
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("exit")
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
		log.Printf("interrupt signal received. shutting down HTTP server with grace period of %v", httpGrace)
		ctx, cancel := context.WithTimeout(context.Background(), httpGrace)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("HTTP server starting to listen on %v", httpAddr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener
		return err
	}
	<-idleConnsClosed
	return nil
}
