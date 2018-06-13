package httpserver

import (
	"net/http"

	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	log = zaplog.Zap("http")
)

// Listen listens http server
func Listen(addr string) {
	log.Info("init", "addr", addr)
	if err := http.ListenAndServe(addr, initHandlers()); err != nil {
		log.Fatal("listen-error", "error", err)
	}
}
