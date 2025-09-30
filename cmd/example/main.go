package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/hudsn/spicyreload/example"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	example.ExampleServer(ctx, 8080)
}
