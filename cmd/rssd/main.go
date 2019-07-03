package main

import (
	"context"
	"flag"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/go-chi/chi"
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
	tmpl, err := template.ParseFiles("template/base.html", "template/index.html")
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, "")
		if err != nil {
			log.WithError(err).Error("rendering template")
		}
	})
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: r,
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
