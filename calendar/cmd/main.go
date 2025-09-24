package main

import (
	"Calendar/internal/app"
	"Calendar/pkg/logger"
	"context"
	"os"
	"os/signal"
)

func main() {
	log := logger.New("env")
	app := app.New(log, "0.0.0.0", "8000")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app.Start()
	<-ctx.Done()
	app.Stop()
}
