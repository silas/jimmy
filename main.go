package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/silas/jimmy/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := cmd.New()

	err := c.Execute()
	if err != nil {
		c.PrintErrln(err.Error())
		os.Exit(1)
	}
}
