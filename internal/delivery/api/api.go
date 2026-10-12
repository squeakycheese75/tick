// Package api serves the application's use cases over HTTP. It is a second front door beside
// the CLI: it translates requests into use case inputs and results into JSON, and holds no
// business rules of its own.
package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

const (
	// maxBodyBytes bounds a request body; a year of trades is well under it.
	maxBodyBytes = 1 << 20
	// maxTradesPerRequest bounds one import so a request cannot hold the write lock for long.
	maxTradesPerRequest = 1000
)

type (
	portfolioLister interface {
		Execute(ctx context.Context) (domain.ListPortfoliosOutput, error)
	}
	portfolioCreator interface {
		Execute(ctx context.Context, in domain.CreatePortfolioUsecaseInput) (*domain.CreatePortfolioUsecaseOutout, error)
	}
	tradesImporter interface {
		Execute(ctx context.Context, in domain.ImportTradesInput) (*domain.ImportTradesOutput, error)
	}
	summaryGetter interface {
		Execute(ctx context.Context, in domain.GetPortfolioSummaryUsecaseInput) (domain.GetPortfolioSummaryUsecaseOutput, error)
	}
	dailyGetter interface {
		Execute(ctx context.Context, in domain.GetDailyReportInput) (domain.GetDailyReportOutput, error)
	}
)

// Deps are the use cases the API exposes.
type Deps struct {
	ListPortfolios  portfolioLister
	CreatePortfolio portfolioCreator
	ImportTrades    tradesImporter
	GetSummary      summaryGetter
	GetDaily        dailyGetter
}

// Options configure the server.
type Options struct {
	// Token is the bearer token every request except /healthz must present. It must not be empty.
	Token string
	// Logger receives one line per request and any internal error. Defaults to discarding.
	Logger *slog.Logger
}

type server struct {
	deps      Deps
	tokenHash [sha256.Size]byte
	log       *slog.Logger
}

// New returns the API handler. It panics if no token is given, because a server holding
// someone's portfolio must never run open by accident.
func New(deps Deps, opts Options) http.Handler {
	if opts.Token == "" {
		panic("api: a bearer token is required")
	}

	log := opts.Logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	s := &server{deps: deps, tokenHash: sha256.Sum256([]byte(opts.Token)), log: log}

	mux := http.NewServeMux()

	for _, r := range s.routes() {
		if r.public {
			mux.HandleFunc(r.pattern, r.handler)
			continue
		}

		mux.Handle(r.pattern, s.authed(r.handler))
	}

	return s.recoverPanics(s.logRequests(s.localHostOnly(securityHeaders(mux))))
}

type route struct {
	pattern string
	handler http.HandlerFunc
	// public routes need no token.
	public bool
}

// routes is the one list of what the API serves; the OpenAPI document is checked against it.
func (s *server) routes() []route {
	return []route{
		{pattern: "GET /healthz", handler: s.health, public: true},
		{pattern: "GET /v1/portfolios", handler: s.listPortfolios},
		{pattern: "POST /v1/portfolios", handler: s.createPortfolio},
		{pattern: "POST /v1/portfolios/{name}/trades", handler: s.importTrades},
		{pattern: "GET /v1/portfolios/{name}/summary", handler: s.summary},
		{pattern: "GET /v1/portfolios/{name}/daily", handler: s.daily},
	}
}

// Routes lists the method and path of everything the API serves, such as "GET /healthz".
func Routes() []string {
	routes := (&server{}).routes()
	patterns := make([]string, 0, len(routes))

	for _, r := range routes {
		patterns = append(patterns, r.pattern)
	}

	return patterns
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// authed requires the bearer token.
func (s *server) authed(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.validToken(r.Header.Get("Authorization")) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="tick"`)
			writeError(w, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
			return
		}

		next(w, r)
	})
}

// validToken compares hashes, so the comparison takes the same time whatever is sent.
func (s *server) validToken(header string) bool {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return false
	}

	got := sha256.Sum256([]byte(strings.TrimSpace(token)))

	return subtle.ConstantTimeCompare(got[:], s.tokenHash[:]) == 1
}

// localHostOnly rejects requests whose Host is not a loopback name. A web page cannot then
// reach the API through a hostname it controls that resolves to 127.0.0.1 (DNS rebinding).
func (s *server) localHostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackHost(r.Host) {
			writeError(w, http.StatusForbidden, "forbidden", "requests must be made to localhost")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isLoopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}

	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests logs method, path, status and time. It never logs headers or bodies, which carry
// the token and the portfolio.
func (s *server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		s.log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", time.Since(start).Milliseconds())
	})
}

func (s *server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				s.log.Error("panic handling request", "method", r.Method, "path", r.URL.Path, "panic", p)
				writeError(w, http.StatusInternalServerError, "internal", "internal error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// decodeJSON reads a JSON request body into v. It insists on JSON, caps the size, and rejects
// unknown fields and trailing data, so a typo in a field name is an error and not a silent no-op.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &requestError{status: http.StatusUnsupportedMediaType, code: "unsupported_media_type", message: "Content-Type must be application/json"}
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return badBody(err)
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return &requestError{status: http.StatusBadRequest, code: "bad_request", message: "the body must be a single JSON value"}
	}

	return nil
}

func badBody(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return &requestError{status: http.StatusRequestEntityTooLarge, code: "too_large", message: "the request body is too large"}
	}

	return &requestError{status: http.StatusBadRequest, code: "bad_request", message: "the body is not valid JSON for this request: " + err.Error()}
}

// requestError is a problem with the request itself, decided before any use case runs.
type requestError struct {
	status  int
	code    string
	message string
}

func (e *requestError) Error() string { return e.message }
