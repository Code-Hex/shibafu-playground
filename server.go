package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type server struct {
	mux *http.ServeMux
	db  store
	log logger
	//cache *gobCache

	// When the executable was last modified. Used for caching headers of compiled assets.
	modtime time.Time

	hostname string
}

func NewServer(hostname string, db store, log logger) *server {
	var modtime time.Time
	execpath, _ := os.Executable()
	if execpath != "" {
		if fi, _ := os.Stat(execpath); fi != nil {
			modtime = fi.ModTime()
		}
	}
	srv := &server{
		mux:      http.NewServeMux(),
		db:       db,
		log:      log,
		modtime:  modtime,
		hostname: hostname,
	}
	srv.initHandlers()
	return srv
}

func (s *server) initHandlers() {
	s.mux.HandleFunc("/", s.editHandler())
	s.mux.HandleFunc("/runtime", s.runtimeHandler())
	s.mux.HandleFunc("/share", s.shareHandler())
	s.mux.HandleFunc("/favicon.ico", faviconHandler())
	s.mux.HandleFunc("/_ah/health", s.healthCheckHandler())

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	s.mux.Handle("/static/", staticHandler)
}

func (s *server) healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}
}

func faviconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-Proto") == "http" {
		r.URL.Scheme = "https"
		r.URL.Host = r.Host
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
	}
	s.log.Printf("request path: %s", r.URL.Path)
	s.mux.ServeHTTP(w, r)
}
