package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/app"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/config"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/httpapi"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/platform/httpserver"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/platform/logger"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store/memory"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store/postgres"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.Env, cfg.LogLevel)
	slog.SetDefault(log)
	if cfg.Env == "production" && cfg.AutoMigrate {
		log.Warn("auto migrate is enabled in production; run migrations as a release step when more than one instance is running")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	users, tokens, tasks, pinger, closeStore, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeStore()

	application, err := app.New(app.Options{
		Config: cfg,
		Logger: log,
		Users:  users,
		Tokens: tokens,
		Tasks:  tasks,
		Pinger: pinger,
	})
	if err != nil {
		return err
	}

	srv := httpserver.New(cfg.HTTPAddr, application.Handler)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("listening", "addr", cfg.HTTPAddr, "env", cfg.Env, "store", cfg.Store)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		log.Info("shutting down")
		return srv.Shutdown(shutdownCtx)
	})
	g.Go(func() error {
		return application.RunCleanup(ctx)
	})

	return g.Wait()
}

func openStore(ctx context.Context, cfg config.Config) (store.UserRepository, store.RefreshTokenRepository, store.TaskRepository, httpapi.Pinger, func(), error) {
	switch cfg.Store {
	case "memory":
		mem := memory.New()
		users, tokens, tasks := memory.Repos(mem)
		return users, tokens, tasks, nil, func() {}, nil
	case "postgres":
		pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, nil, nil, func() {}, err
		}
		pingCtx, cancel := context.WithTimeout(ctx, cfg.RequestTimeout)
		defer cancel()
		if err := pool.Ping(pingCtx); err != nil {
			pool.Close()
			return nil, nil, nil, nil, func() {}, fmt.Errorf("database ping: %w", err)
		}
		if cfg.AutoMigrate {
			if err := postgres.Migrate(ctx, pool); err != nil {
				pool.Close()
				return nil, nil, nil, nil, func() {}, err
			}
		}
		users, tokens, tasks := postgres.Repos(postgres.NewRepository(pool))
		return users, tokens, tasks, pool, pool.Close, nil
	default:
		return nil, nil, nil, nil, func() {}, fmt.Errorf("unknown store %q", cfg.Store)
	}
}

func runHealthcheck() int {
	addr := os.Getenv("API_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(healthURL(addr))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		return 1
	}
	return 0
}

func healthURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr + "/healthz"
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr + "/healthz"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/healthz"
}
