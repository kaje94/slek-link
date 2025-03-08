package main

import (
	"log/slog"
	"os"

	"github.com/kaje94/slek-link/webapp/internal"
)

func main() {
	if err := internal.RunServer(); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}
