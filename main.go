package main

import (
	hs "github.com/samasno/http-server/pkg/http"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mux := hs.NewHandler()
	mux.HandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("home page"))
	})

	srv := &hs.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	k := make(chan os.Signal, 5)
	signal.Notify(k, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func(k chan os.Signal, srv *hs.Server) {
		<-k
		srv.Close()
	}(k, srv)

	if err := srv.ListenAndServeHttp(); err != nil {
		log.Fatal(err.Error())
	}
}
