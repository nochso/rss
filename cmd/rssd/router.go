package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func newRouter() http.Handler {
	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "index.html", "")
	})
	return r
}
