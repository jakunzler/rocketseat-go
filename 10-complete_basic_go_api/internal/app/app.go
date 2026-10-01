package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/auth"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/config"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/httpapi"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

type Options struct {
	Config config.Config
	Logger *slog.Logger
	Users  store.UserRepository
	Tokens store.RefreshTokenRepository
	Tasks  store.TaskRepository
	Pinger httpapi.Pinger
	Hasher service.PasswordHasher
	Now    func() time.Time
}

type App struct {
	Handler http.Handler
	cleanup func(context.Context) error
}

func New(opts Options) (*App, error) {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Hasher == nil {
		opts.Hasher = auth.Argon2{}
	}

	issuer, err := auth.NewIssuer([]byte(opts.Config.JWTSecret), opts.Config.AccessTokenTTL, opts.Config.JWTIssuer, opts.Now)
	if err != nil {
		return nil, err
	}
	authService, err := service.NewAuth(opts.Users, opts.Tokens, issuer, opts.Hasher, opts.Config.RefreshTokenTTL, opts.Now, opts.Logger)
	if err != nil {
		return nil, err
	}

	handler := httpapi.New(httpapi.Options{
		Logger:             opts.Logger,
		Auth:               authService,
		Tasks:              service.NewTasks(opts.Tasks, opts.Now),
		Issuer:             issuer,
		Pinger:             opts.Pinger,
		CORSOrigins:        opts.Config.CORSOrigins,
		RateLimitRPS:       opts.Config.RateLimitRPS,
		RateLimitBurst:     opts.Config.RateLimitBurst,
		AuthRateLimitRPS:   opts.Config.AuthRateLimitRPS,
		AuthRateLimitBurst: opts.Config.AuthRateLimitBurst,
		RequestTimeout:     opts.Config.RequestTimeout,
		HSTS:               opts.Config.Env == "production",
		TrustProxy:         opts.Config.TrustProxy,
	})

	return &App{
		Handler: handler,
		cleanup: func(ctx context.Context) error {
			return authService.RunCleanup(ctx, opts.Config.CleanupInterval)
		},
	}, nil
}

func (a *App) RunCleanup(ctx context.Context) error {
	if a.cleanup == nil {
		<-ctx.Done()
		return nil
	}
	return a.cleanup(ctx)
}
