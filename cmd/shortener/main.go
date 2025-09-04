package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

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

func main() {
	buildOut()

	cfg, err := config.CreateConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)

	store, err := config.CreateStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	auth := auth.NewAuthConfig()

	h := handlers.CreateHandle(cfg, store, auth)

	mux := api.InitRoute(&h)

	if cfg.HTTPSMode == "true" {
		cert := `.\internal\cert\server.crt`
		key := `.\internal\cert\server.key`
		err = http.ListenAndServeTLS(cfg.ServerURL, cert, key, mux)
	} else {
		fmt.Println(cfg.ServerURL)
		err = http.ListenAndServe(cfg.ServerURL, mux)
	}

	if err != nil {
		log.Fatal(err)
	}
}
