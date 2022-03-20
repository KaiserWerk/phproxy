package handler

import (
	"log"
	"net/http"

	"github.com/KaiserWerk/phproxy/internal/config"
)

type Base struct {
	Config *config.AppConfig
	Logger *log.Logger
}

func (b *Base) Handler(w http.ResponseWriter, r *http.Request) {

}
