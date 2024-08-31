package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/carlmjohnson/requests"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()
	short, _ := strconv.ParseBool(flag.CommandLine.Lookup("test.short").Value.String())
	if short {
		log.Println("skipping initial setup for short tests")
		os.Exit(m.Run())
		return
	}

	port := port()
	go func() {
		err := run(ctx, os.Stdout, []string{"realworld", "--port", port})
		if err != nil {
			fmt.Printf("failed to run in test %s\n", err)
		}
	}()

	waitForHealthy(ctx, 2*time.Second, endpoint()+"/health")
	os.Exit(m.Run())
}

func endpoint() string {
	return "http://localhost:" + port()
}

func port() string {
	_portOnce.Do(func() {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatalf("Failed to get a free port: %v", err)
		}
		defer listener.Close()

		addr := listener.Addr().(*net.TCPAddr)
		_port = strconv.Itoa(addr.Port)
		log.Printf("Using port %s", _port)
	})
	return _port
}

var (
	_portOnce sync.Once
	_port     string
)

func waitForHealthy(ctx context.Context, timeout time.Duration, endpoint string) {
	startTime := time.Now()
	for {
		err := requests.URL(endpoint).Fetch(ctx)
		if err == nil {
			log.Println("endpoint is ready")
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			if timeout <= time.Since(startTime) {
				log.Fatalf("timeout reached white waitForHealthy")
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
