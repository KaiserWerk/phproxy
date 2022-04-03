package handler

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/KaiserWerk/phproxy/internal/config"
)

type Base struct {
	Config *config.AppConfig
	Logger *log.Logger
}

func (b *Base) Handler(proxy *httputil.ReverseProxy, targetHost string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func (b *Base) CustomRouteHandler(proxy *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO

		proxy.ServeHTTP(w, r)
	}
}
