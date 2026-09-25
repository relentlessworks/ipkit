package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/relentlessworks/ipkit/internal/api"
	"github.com/relentlessworks/ipkit/internal/config"
)

func main() {
	cfg := config.Load()

	// Generate a random secret if none provided
	if cfg.Secret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("failed to generate secret: %v", err)
		}
		cfg.Secret = hex.EncodeToString(b)
	}

	handler := api.NewHandler(cfg.Secret)
	mux := handler.Routes()

	log.Printf("ipkit listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
