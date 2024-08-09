package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/raeperd/realworld"
	"github.com/raeperd/realworld/internal/inmemory"
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

	userService := realworld.NewUserService(
		inmemory.NewUserRepository(),
	)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", strconv.Itoa(int(port))),
		Handler: newServer(userService),
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	// server blocks here until os.Interrput
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
