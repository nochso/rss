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
	dbFile    = "rss.sqlite3"
	httpAddr  = ":8080"
	httpGrace = time.Second * 10
	templates map[string]*template.Template
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(logfmt.Default)
	flag.StringVar(&dbFile, "db", dbFile, "sqlite3 db file")
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
	var err error
	templates, err = parseTemplates()
	if err != nil {
		return err
	}

	db, err := openDB(dbFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.WithError(err).Error("closing db")
			return
		}
		log.Info("db closed")
	}()

	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "index.html", "")
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

func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, ok := templates[tmpl]
	if !ok {
		log.WithField("template", tmpl).Error("template does not exist")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.Execute(w, data)
	if err != nil {
		log.WithField("template", tmpl).WithError(err).Error("rendering template")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// parseTemplates returns a map of parsed templates with template names for keys.
func parseTemplates() (map[string]*template.Template, error) {
	// template name to required template files
	paths := map[string][]string{
		"index.html": {"template/base.html", "template/index.html"},
	}
	tmpl := make(map[string]*template.Template, len(paths))
	var err error
	for name, files := range paths {
		tmpl[name], err = template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}
