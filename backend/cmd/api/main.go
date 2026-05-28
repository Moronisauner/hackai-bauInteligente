package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moronisauner/hackai/backend/internal/config"
	"github.com/moronisauner/hackai/backend/internal/db"
	"github.com/moronisauner/hackai/backend/internal/httpapi"

	_ "github.com/moronisauner/hackai/backend/docs" // docs gerados pelo swag
)

//	@title			Centralizador de Saldo (POC) — API
//	@version		0.1
//	@description	Backend da POC do centralizador de saldo. Toda regra temporal usa POC_REFERENCE_DATE.
//	@BasePath		/
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(pingCtx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Contexto cancelado em SIGINT/SIGTERM para shutdown gracioso.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: httpapi.New(pool, cfg),
	}

	go func() {
		log.Printf("hackai api listening on http://localhost%s (POCReferenceDate=%s)",
			cfg.HTTPPort, cfg.POCReferenceDate.Format("2006-01-02"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
