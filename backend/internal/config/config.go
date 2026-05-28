// Package config carrega a configuração da app a partir de variáveis de
// ambiente. Falha explicitamente quando uma variável obrigatória está ausente
// ou malformada — nunca chama log.Fatal aqui; quem decide é o cmd/api.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config agrega toda a configuração de runtime do backend.
type Config struct {
	DatabaseURL      string    // obrigatória
	POCReferenceDate time.Time // obrigatória, formato YYYY-MM-DD
	HTTPPort         string    // default ":8080"
}

// Load lê a configuração do ambiente. Retorna erro descritivo (prefixado com
// "config:") se alguma obrigatória faltar ou não puder ser parseada.
func Load() (Config, error) {
	var cfg Config

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}

	rawDate := os.Getenv("POC_REFERENCE_DATE")
	if rawDate == "" {
		return Config{}, fmt.Errorf("config: POC_REFERENCE_DATE is required (format YYYY-MM-DD)")
	}
	refDate, err := time.Parse("2006-01-02", rawDate)
	if err != nil {
		return Config{}, fmt.Errorf("config: POC_REFERENCE_DATE %q is invalid, expected YYYY-MM-DD: %w", rawDate, err)
	}
	cfg.POCReferenceDate = refDate

	cfg.HTTPPort = os.Getenv("HTTP_PORT")
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = ":8080"
	}

	return cfg, nil
}
