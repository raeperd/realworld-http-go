package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	var port uint
	fs := flag.NewFlagSet("realworld", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.UintVar(&port, "port", 8080, "port to use in http server")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tick := time.Tick(time.Second)
	var seconds uint
	for {
		select {
		case <-tick:
			seconds += 1
			fmt.Fprintf(w, "service running for %v seconds\n", seconds)
		case <-ctx.Done():
			fmt.Fprintf(w, "\nservice exit after %v seconds with signal", seconds)
			return nil
		}
	}
}
