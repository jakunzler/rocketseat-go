package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/auth"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
	"golang.org/x/time/rate"
)

type API struct {
	log            *slog.Logger
	authn          *service.Auth
	tasks          *service.Tasks
	issuer         *auth.Issuer
	pinger         Pinger
	metrics        *metrics
	hsts           bool
	requestTimeout time.Duration
	corsOrigins    []string
	limit          *ipLimiter
	authLimit      *ipLimiter
}

type Options struct {
	Logger             *slog.Logger
	Auth               *service.Auth
	Tasks              *service.Tasks
	Issuer             *auth.Issuer
	Pinger             Pinger
	CORSOrigins        []string
	RateLimitRPS       float64
	RateLimitBurst     int
	AuthRateLimitRPS   float64
	AuthRateLimitBurst int
	RequestTimeout     time.Duration
	HSTS               bool
	TrustProxy         bool
}

func New(opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = 15 * time.Second
	}
	api := &API{
		log:            opts.Logger,
		authn:          opts.Auth,
		tasks:          opts.Tasks,
		issuer:         opts.Issuer,
		pinger:         opts.Pinger,
		metrics:        newMetrics(),
		hsts:           opts.HSTS,
		requestTimeout: opts.RequestTimeout,
		corsOrigins:    opts.CORSOrigins,
		limit:          newIPLimiter(rate.Limit(opts.RateLimitRPS), opts.RateLimitBurst, skipInfra),
		authLimit:      newIPLimiter(rate.Limit(opts.AuthRateLimitRPS), opts.AuthRateLimitBurst, nil),
	}
	return newRouter(api, opts.TrustProxy)
}
