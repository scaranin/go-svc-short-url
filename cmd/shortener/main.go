package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/api"
	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func buildOut() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)

	if buildDate == "" {
		buildDate = "N/A"
	}
	fmt.Printf("Build date: %s\n", buildDate)

	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func startServer(cfg *config.ShortenerConfig, router *chi.Mux) error {
	var err error
	server := http.Server{
		Addr:    cfg.ServerURL,
		Handler: router,
	}

	go func() {
		if cfg.HTTPSMode == "true" {
			cert := `.\internal\cert\server.crt`
			key := `.\internal\cert\server.key`
			err = http.ListenAndServeTLS(cfg.ServerURL, cert, key, router)
		} else {
			fmt.Println(cfg.ServerURL)
			err = http.ListenAndServe(cfg.ServerURL, router)
		}

		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)

	if err != nil {
		return err
	} else {
		log.Println("Server is stopped!")
	}

	return err
}

func main() {
	buildOut()

	cfg, err := config.CreateConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := config.CreateStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	auth := auth.NewAuthConfig()

	h := handlers.CreateHandle(cfg, store, auth)

	mux := api.InitRoute(&h)

	startServer(&cfg, mux)

	if err != nil {
		log.Fatal(err)
	}
}
