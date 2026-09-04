// Command server starts the OSINT research backend.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akeelnazir/osint-app/backend/internal/config"
	"github.com/akeelnazir/osint-app/backend/internal/server"

	_ "github.com/lib/pq" // register the postgres driver for database/sql
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}
	defer srv.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
	log.Println("server shut down cleanly")
}
