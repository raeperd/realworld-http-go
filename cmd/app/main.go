package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/carlmjohnson/versioninfo"
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
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var port uint
	fs := flag.NewFlagSet("realworld", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.UintVar(&port, "port", 8080, "port to use in http server")
	versioninfo.AddFlag(fs)
	fs.Usage = func() {
		fmt.Fprintf(w, "Usage of %s:\n", fs.Name())
		fmt.Fprintf(w, "This is a simple program that greets a person.\n\n")
		fmt.Fprintf(w, "Flags:\n")
		fs.PrintDefaults()
	}
	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	userService := realworld.NewUserService(
		inmemory.NewUserRepository(),
	)

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(int(port)),
		Handler: newServer(userService),
	}
	go func() {
		// TODO: Use slog
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	// NOTE: server blocks here until os.Interrput
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
