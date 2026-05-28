package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

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

	printBootBanner(cfg.POCReferenceDate)

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

// printBootBanner imprime, antes dos logs estruturados, um banner em destaque
// com a POC_REFERENCE_DATE em uso. Toda regra de negócio usa essa data e NÃO
// time.Now() (PRD §8) — o banner existe pra deixar isso impossível de ignorar.
// A largura é calculada a partir do conteúdo, então acomoda qualquer data.
func printBootBanner(refDate time.Time) {
	lines := []string{
		"POC_REFERENCE_DATE = " + refDate.Format("2006-01-02"),
		"Toda regra de negócio usa essa data,",
		"NÃO time.Now()",
	}

	// Largura interna = maior linha (em runes, não bytes, por causa dos acentos)
	// + 2 espaços de padding de cada lado.
	width := 0
	for _, l := range lines {
		if n := utf8.RuneCountInString(l); n > width {
			width = n
		}
	}
	width += 2

	top := "╔" + strings.Repeat("═", width) + "╗"
	bottom := "╚" + strings.Repeat("═", width) + "╝"

	fmt.Println(top)
	for _, l := range lines {
		pad := width - 1 - utf8.RuneCountInString(l)
		fmt.Printf("║ %s%s║\n", l, strings.Repeat(" ", pad))
	}
	fmt.Println(bottom)
}
