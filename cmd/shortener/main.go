package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/api"
	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/gen"
	grpsShortener "github.com/scaranin/go-svc-short-url/internal/grpc"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/models"
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

func startServer(cfg *config.ShortenerConfig, router *chi.Mux) (*http.Server, error) {
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
			fmt.Println("Starting REST server on:", cfg.ServerURL)
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal("REST server error:", err)
		}
	}()

	return &server, nil
}

func startServerGRPC(cfg *config.ShortenerConfig, store models.Storage, auth *auth.AuthConfig) error {
	grpcAuthService := grpsShortener.NewAuthService(auth)
	grpcServer := grpsShortener.NewGRPCServer(store, cfg.BaseURL, grpcAuthService)

	server := grpc.NewServer()
	gen.RegisterShortenerServiceServer(server, grpcServer)

	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return err
	}

	go func() {
		fmt.Println("Starting gRPC server on:", cfg.GRPCAddress)
		if err := server.Serve(listener); err != nil {
			log.Fatal("gRPC server error:", err)
		}
	}()

	return nil
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

	server, err := startServer(&cfg, mux)
	if err != nil {
		log.Fatal("Failed to start REST server:", err)
	}

	err = startServerGRPC(&cfg, store, &auth)
	if err != nil {
		log.Fatal("Failed to start gRPC server:", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Both servers are running. Press Ctrl+C to stop.")

	<-signalChan
	fmt.Println("\nShutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	fmt.Println("Servers stopped!")
}
