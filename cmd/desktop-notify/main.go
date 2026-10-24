package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pouyade/go-mock-sms"
	"github.com/pouyade/go-mock-sms/internal/httpapi"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 4466
)

type config struct {
	host        string
	port        int
	title       string
	showVersion bool
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(2)
	}
	if cfg.showVersion {
		fmt.Printf("desktop-notify %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if !isLoopbackHost(cfg.host) {
		logger.Warn("HTTP server is listening beyond loopback; protect the endpoint with network controls",
			"host", cfg.host)
	}

	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port)),
		Handler:           httpapi.NewMux(notify.New(), cfg.title, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-shutdownContext.Done()
		timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(timeoutContext); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("desktop notification service listening",
		"address", server.Addr,
		"version", version,
	)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server stopped", "error", err)
		os.Exit(1)
	}
}

func parseConfig(args []string) (config, error) {
	cfg := config{
		host:  envOrDefault("DESKTOP_NOTIFY_HOST", defaultHost),
		port:  envIntOrDefault("DESKTOP_NOTIFY_PORT", defaultPort),
		title: envOrDefault("DESKTOP_NOTIFY_TITLE", notify.DefaultTitle),
	}
	flags := flag.NewFlagSet("desktop-notify", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&cfg.host, "host", cfg.host, "HTTP listen host")
	flags.IntVar(&cfg.port, "port", cfg.port, "HTTP listen port")
	flags.StringVar(&cfg.title, "title", cfg.title, "default notification title")
	flags.BoolVar(&cfg.showVersion, "version", false, "print version and exit")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if flags.NArg() > 0 {
		return config{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if cfg.port < 1 || cfg.port > 65535 {
		return config{}, fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.host) == "" {
		return config{}, fmt.Errorf("host must not be empty")
	}
	if strings.TrimSpace(cfg.title) == "" {
		cfg.title = notify.DefaultTitle
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
