package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/zangdale/httplog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

//go:embed public
var web embed.FS

func main() {
	fsys, err := fs.Sub(web, "public")
	if err != nil {
		log.Fatal().Err(err).Msg("FS sub public path")
		panic("remove public path")
	}

	fi, err := fs.Stat(fsys, "index1.html")
	if err != nil || fi.Size() < 1 {
		log.Fatal().Err(err).Msg("stat public index1.html")
		panic("no index1.html")
	}
	http2.VerboseLogs = true

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(fsys)))
	mux.HandleFunc("/events", handevent)
	mux.HandleFunc("/http", httplog.GetLatestEvents)
	mux.HandleFunc("/update", httplog.PushEvent)

	err = http.ListenAndServe(":80", h2c.NewHandler(mux, &http2.Server{}))
	if err != nil {
		log.Fatal().Err(err).Msg("http listen and serve")
	}
}

func handevent(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if req.Method == http.MethodPost {
		w.Header().Add("content-type", "application/json; charset=utf-8")
		httplog.PushEvent(w, req)
		return
	}

	if req.URL.Query().Get("type") == "http" {
		w.Header().Add("content-type", "application/json; charset=utf-8")
		httplog.GetLatestEvents(w, req)
		return
	}
	httplog.EventSource(w, req)
}
