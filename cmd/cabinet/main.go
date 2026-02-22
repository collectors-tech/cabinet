package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/collectors-tech/cabinet/internal/app"
	"github.com/collectors-tech/cabinet/internal/config"
)

func main() {
	cfg := config.Load()
	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		log.Fatalf("runtime failed: %v", err)
	}
}
