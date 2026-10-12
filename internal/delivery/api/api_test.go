package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/delivery/api"
	"github.com/squeakycheese75/tick/internal/domain"
)

const token = "s3cret-token"

// fakes records what the handlers pass to the use cases and returns canned results.
type fakes struct {
	listOut domain.ListPortfoliosOutput
	listErr error

	createIn  domain.CreatePortfolioUsecaseInput
	createOut *domain.CreatePortfolioUsecaseOutout
	createErr error

	tradesIn  domain.ImportTradesInput
	tradesOut *domain.ImportTradesOutput
	tradesErr error

	summaryIn  domain.GetPortfolioSummaryUsecaseInput
	summaryOut domain.GetPortfolioSummaryUsecaseOutput
	summaryErr error

	dailyIn  domain.GetDailyReportInput
	dailyOut domain.GetDailyReportOutput
	dailyErr error
	dailyRun bool

	panicOnList bool
}

func (f *fakes) lister() lister           { return lister{f} }
func (f *fakes) creator() creator         { return creator{f} }
func (f *fakes) importer() importer       { return importer{f} }
func (f *fakes) summaryGetter() summaryFn { return summaryFn{f} }
func (f *fakes) dailyGetter() dailyFn     { return dailyFn{f} }

type lister struct{ f *fakes }

func (l lister) Execute(context.Context) (domain.ListPortfoliosOutput, error) {
	if l.f.panicOnList {
		panic("kaboom")
	}
	return l.f.listOut, l.f.listErr
}

type creator struct{ f *fakes }

func (c creator) Execute(_ context.Context, in domain.CreatePortfolioUsecaseInput) (*domain.CreatePortfolioUsecaseOutout, error) {
	c.f.createIn = in
	return c.f.createOut, c.f.createErr
}

type importer struct{ f *fakes }

func (i importer) Execute(_ context.Context, in domain.ImportTradesInput) (*domain.ImportTradesOutput, error) {
	i.f.tradesIn = in
	return i.f.tradesOut, i.f.tradesErr
}

type summaryFn struct{ f *fakes }

func (s summaryFn) Execute(_ context.Context, in domain.GetPortfolioSummaryUsecaseInput) (domain.GetPortfolioSummaryUsecaseOutput, error) {
	s.f.summaryIn = in
	return s.f.summaryOut, s.f.summaryErr
}

type dailyFn struct{ f *fakes }

func (d dailyFn) Execute(_ context.Context, in domain.GetDailyReportInput) (domain.GetDailyReportOutput, error) {
	d.f.dailyIn, d.f.dailyRun = in, true
	return d.f.dailyOut, d.f.dailyErr
}

func newFakes() *fakes {
	return &fakes{
		createOut: &domain.CreatePortfolioUsecaseOutout{PortfolioName: "trending", BaseCurrency: "EUR"},
		tradesOut: &domain.ImportTradesOutput{PortfolioName: "trending", Applied: 2, Skipped: 1},
	}
}

type harness struct {
	f       *fakes
	handler http.Handler
	logs    *bytes.Buffer
}

func newHarness(t *testing.T, mutate func(*fakes)) *harness {
	t.Helper()

	f := newFakes()
	if mutate != nil {
		mutate(f)
	}

	logs := &bytes.Buffer{}
	h := api.New(api.Deps{
		ListPortfolios:  f.lister(),
		CreatePortfolio: f.creator(),
		ImportTrades:    f.importer(),
		GetSummary:      f.summaryGetter(),
		GetDaily:        f.dailyGetter(),
	}, api.Options{Token: token, Logger: slog.New(slog.NewTextHandler(logs, nil))})

	return &harness{f: f, handler: h, logs: logs}
}

type request struct {
	method, path, body string
	header             map[string]string
	host               string
	noAuth             bool
}

func (h *harness) do(r request) *httptest.ResponseRecorder {
	req := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	req.Host = "127.0.0.1:8080"
	if r.host != "" {
		req.Host = r.host
	}
	if !r.noAuth {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if r.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range r.header {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)

	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("body is not the expected JSON: %v\n%s", err, rec.Body.String())
	}

	return v
}

type errResp struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestNew_RefusesToRunWithoutAToken(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("an API with no token must not be constructible")
		}
	}()

	api.New(api.Deps{}, api.Options{})
}

func TestAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header map[string]string
		noAuth bool
		want   int
	}{
		{"no header", nil, true, http.StatusUnauthorized},
		{"wrong token", map[string]string{"Authorization": "Bearer nope"}, true, http.StatusUnauthorized},
		{"token without scheme", map[string]string{"Authorization": token}, true, http.StatusUnauthorized},
		{"wrong scheme", map[string]string{"Authorization": "Basic " + token}, true, http.StatusUnauthorized},
		{"empty bearer", map[string]string{"Authorization": "Bearer "}, true, http.StatusUnauthorized},
		{"right token", nil, false, http.StatusOK},
		{"scheme is case insensitive", map[string]string{"Authorization": "bearer " + token}, true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios", header: tt.header, noAuth: tt.noAuth})

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if tt.want == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") == "" {
				t.Fatal("a 401 should say how to authenticate")
			}
		})
	}
}

func TestEveryRouteExceptHealthRequiresAuth(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	routes := []struct{ method, path string }{
		{"GET", "/v1/portfolios"},
		{"POST", "/v1/portfolios"},
		{"POST", "/v1/portfolios/p/trades"},
		{"GET", "/v1/portfolios/p/summary"},
		{"GET", "/v1/portfolios/p/daily"},
	}

	for _, r := range routes {
		if rec := h.do(request{method: r.method, path: r.path, noAuth: true}); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a token = %d, want 401", r.method, r.path, rec.Code)
		}
	}

	rec := h.do(request{method: "GET", path: "/healthz", noAuth: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("/healthz = %d, want 200 without a token", rec.Code)
	}
}

func TestUnknownRoutesAndMethods(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)

	if rec := h.do(request{method: "GET", path: "/v1/nothing"}); rec.Code != http.StatusNotFound {
		t.Errorf("unknown path = %d, want 404", rec.Code)
	}
	if rec := h.do(request{method: "DELETE", path: "/v1/portfolios"}); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong method = %d, want 405", rec.Code)
	}
}

func TestOnlyLoopbackHostsAreServed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host string
		want int
	}{
		{"127.0.0.1:8080", http.StatusOK},
		{"localhost:8080", http.StatusOK},
		{"LOCALHOST", http.StatusOK},
		{"[::1]:8080", http.StatusOK},
		{"evil.example.com", http.StatusForbidden},
		{"evil.example.com:8080", http.StatusForbidden},
		{"192.168.1.20:8080", http.StatusForbidden},
		{"localhost.evil.com", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			t.Parallel()

			rec := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios", host: tt.host})
			if rec.Code != tt.want {
				t.Fatalf("Host %q = %d, want %d", tt.host, rec.Code, tt.want)
			}
		})
	}
}

func TestResponsesAreNotCachedOrSniffed(t *testing.T) {
	t.Parallel()

	rec := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios"})

	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("headers = %v", rec.Header())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestListPortfolios(t *testing.T) {
	t.Parallel()

	h := newHarness(t, func(f *fakes) {
		f.listOut = domain.ListPortfoliosOutput{Portfolios: []domain.PortfolioListItem{
			{Name: "main", BaseCurrency: "EUR", PositionCount: 0},
			{Name: "trending", BaseCurrency: "EUR", PositionCount: 5},
		}}
	})

	rec := h.do(request{method: "GET", path: "/v1/portfolios"})
	got := decode[struct {
		Portfolios []map[string]any `json:"portfolios"`
	}](t, rec)

	if rec.Code != http.StatusOK || len(got.Portfolios) != 2 || got.Portfolios[1]["name"] != "trending" || got.Portfolios[1]["positionCount"] != float64(5) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestListPortfolios_EmptyIsAnEmptyArrayNotNull(t *testing.T) {
	t.Parallel()

	rec := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios"})

	if !strings.Contains(rec.Body.String(), `"portfolios":[]`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestCreatePortfolio(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	rec := h.do(request{method: "POST", path: "/v1/portfolios", body: `{"name":"  trending ","baseCurrency":"eur"}`})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if h.f.createIn.PortfolioName != "trending" || h.f.createIn.BaseCurrency != "EUR" {
		t.Fatalf("use case got %+v, want trimmed name and upper-cased currency", h.f.createIn)
	}
	if got := decode[map[string]string](t, rec); got["name"] != "trending" || got["baseCurrency"] != "EUR" {
		t.Fatalf("body = %v", got)
	}
}

func TestImportTrades(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	body := `{"trades":[
		{"ref":"r1","date":"2026-10-12","symbol":"MRNA","side":"buy","quantity":0.995,"price":201,"currency":"EUR","fee":1,"fxRate":1.12,"exchange":"NASDAQ","instrumentType":"equity","quoteCurrency":"USD"},
		{"date":"2026-10-13T09:30:00Z","symbol":"DELL","side":"sell","quantity":1,"price":5,"currency":"USD"}
	]}`

	rec := h.do(request{method: "POST", path: "/v1/portfolios/trending/trades", body: body})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if got := decode[map[string]any](t, rec); got["portfolio"] != "trending" || got["applied"] != float64(2) || got["skipped"] != float64(1) {
		t.Fatalf("body = %v", got)
	}

	in := h.f.tradesIn
	if in.PortfolioName != "trending" || len(in.Trades) != 2 {
		t.Fatalf("use case got %+v", in)
	}

	first := in.Trades[0]
	want := domain.TradeInput{
		Ref: "r1", Date: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Symbol: "MRNA", Side: domain.TradeSideBuy,
		Quantity: 0.995, Price: 201, Currency: "EUR", Fee: 1, FXRate: 1.12, Exchange: "NASDAQ", InstrumentType: "equity", QuoteCurrency: "USD",
	}
	if first != want {
		t.Fatalf("first trade = %+v, want %+v", first, want)
	}
	if !in.Trades[1].Date.Equal(time.Date(2026, 10, 13, 9, 30, 0, 0, time.UTC)) {
		t.Fatalf("an RFC 3339 date should be accepted: %v", in.Trades[1].Date)
	}
}

func TestImportTrades_PortfolioNameWithEscapedCharacters(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	h.do(request{method: "POST", path: "/v1/portfolios/my%20fund/trades", body: `{"trades":[]}`})

	if h.f.tradesIn.PortfolioName != "my fund" {
		t.Fatalf("portfolio name = %q, want it URL-decoded", h.f.tradesIn.PortfolioName)
	}
}

func TestImportTrades_RequestProblems(t *testing.T) {
	t.Parallel()

	tooMany := `{"trades":[` + strings.Repeat(`{"date":"2026-10-12","symbol":"A","side":"buy","quantity":1,"price":1,"currency":"USD"},`, 1000) +
		`{"date":"2026-10-12","symbol":"A","side":"buy","quantity":1,"price":1,"currency":"USD"}]}`

	tests := []struct {
		name       string
		body       string
		header     map[string]string
		wantStatus int
		wantCode   string
	}{
		{"not JSON", `{nope`, nil, http.StatusBadRequest, "bad_request"},
		{"unknown field", `{"trades":[],"extra":1}`, nil, http.StatusBadRequest, "bad_request"},
		{"misspelled field is an error, not a silent skip", `{"trades":[{"date":"2026-10-12","symbl":"A"}]}`, nil, http.StatusBadRequest, "bad_request"},
		{"wrong type", `{"trades":[{"quantity":"lots"}]}`, nil, http.StatusBadRequest, "bad_request"},
		{"two JSON values", `{"trades":[]} {"trades":[]}`, nil, http.StatusBadRequest, "bad_request"},
		{"trailing garbage", `{"trades":[]} nope`, nil, http.StatusBadRequest, "bad_request"},
		{"wrong content type", `{"trades":[]}`, map[string]string{"Content-Type": "text/plain"}, http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"no content type", `{"trades":[]}`, map[string]string{"Content-Type": ""}, http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"content type with charset is fine", `{"trades":[]}`, map[string]string{"Content-Type": "application/json; charset=utf-8"}, http.StatusOK, ""},
		{"bad date", `{"trades":[{"date":"yesterday","symbol":"A","side":"buy","quantity":1,"price":1,"currency":"USD"}]}`, nil, http.StatusUnprocessableEntity, "invalid"},
		{"too many trades", tooMany, nil, http.StatusUnprocessableEntity, "invalid"},
		{"body too large", `{"trades":[],"padding":"` + strings.Repeat("x", 2<<20) + `"}`, nil, http.StatusRequestEntityTooLarge, "too_large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t, nil)
			rec := h.do(request{method: "POST", path: "/v1/portfolios/p/trades", body: tt.body, header: tt.header})

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantCode != "" {
				if got := decode[errResp](t, rec).Error.Code; got != tt.wantCode {
					t.Fatalf("code = %q, want %q", got, tt.wantCode)
				}
			}
		})
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	h := newHarness(t, func(f *fakes) {
		f.summaryOut = domain.GetPortfolioSummaryUsecaseOutput{Report: domain.PortfoloSummaryReport{
			PortfolioName: "trending", BaseCurrency: "EUR", PricesAsOf: asOf,
			TotalValue: 1006.62, TotalCost: 1005, TotalPnL: 1.62, TotalPnLPct: 0.0016, TotalAssetPnL: 1.15, TotalFXPnL: 0.47,
			Positions: []domain.SummaryPosition{
				{Symbol: "SNDK", Quantity: 0.14, QuotedPrice: 1581.82, InstrumentCurrency: "USD", MarketValueBase: 202.98, CostBasisBase: 201, UnrealizedPnL: 1.98, UnrealizedPnLPct: 0.0099, AssetPnL: 1.89, FXPnL: 0.09, FXKnown: true, Weight: 0.2},
				{Symbol: "OLD", Quantity: 1, QuotedPrice: 5, InstrumentCurrency: "USD", MarketValueBase: 4, CostBasisBase: 3, UnrealizedPnL: 1, UnrealizedPnLPct: 0.33, Weight: 0.1},
			},
		}}
	})

	rec := h.do(request{method: "GET", path: "/v1/portfolios/trending/summary"})
	if rec.Code != http.StatusOK || h.f.summaryIn.PortfolioName != "trending" {
		t.Fatalf("status %d, use case got %+v", rec.Code, h.f.summaryIn)
	}

	got := decode[map[string]any](t, rec)
	if got["portfolio"] != "trending" || got["totalPnLRatio"] != 0.0016 || got["totalFXPnL"] != 0.47 || got["pricesAsOf"] != "2026-10-09T20:00:00Z" {
		t.Fatalf("body = %v", got)
	}

	positions := got["positions"].([]any)
	tracked, untracked := positions[0].(map[string]any), positions[1].(map[string]any)
	if tracked["assetPnL"] != 1.89 || tracked["fxPnL"] != 0.09 {
		t.Fatalf("tracked position = %v", tracked)
	}
	if untracked["assetPnL"] != nil || untracked["fxPnL"] != nil {
		t.Fatalf("a position without a recorded purchase rate should have null asset and FX PnL, got %v", untracked)
	}
}

func TestSummary_OmitsPricesAsOfWhenUnknown(t *testing.T) {
	t.Parallel()

	rec := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios/p/summary"})

	if strings.Contains(rec.Body.String(), "pricesAsOf") || !strings.Contains(rec.Body.String(), `"positions":[]`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestDaily(t *testing.T) {
	t.Parallel()

	h := newHarness(t, func(f *fakes) {
		f.dailyOut = domain.GetDailyReportOutput{
			AISummary: "all quiet",
			DailyReport: domain.DailyReport{
				Portfolio:       domain.PortfolioSummary{Name: "trending", BaseCurrency: "EUR", TotalValue: 1006.62},
				TopHoldings:     domain.HoldingSummary{Holdings: []domain.Holding{{Symbol: "MRNA", Quantity: 0.99, Weight: 0.2, MarketValueBase: 199.67, QuotedPrice: 225, PriceCurrency: "USD", ChangePercent: 14.21, ChangeAbsolute: 27.86}}},
				Risk:            domain.RiskSummary{LargestPosition: "SNDK", LargestWeight: 0.2016, Top3Concentration: 0.6031},
				News:            []domain.NewsSummary{{Ticker: "MRNA", Headlines: []domain.NewsHeadline{{Title: "Moderna does a thing", URL: "https://example.com/a"}}}},
				Attention:       []string{"SNDK is 20.16% of the portfolio"},
				Targets:         []domain.TargetStatus{{Symbol: "MRNA", Type: domain.TargetTypeStopLoss, CurrentPrice: 225, TargetPrice: 200, Currency: "USD", DistancePct: 0.11}},
				ValuationIssues: []domain.ValuationIssue{{Symbol: "FUND", Type: "missing_price", Message: "no price", Quantity: 3}},
			},
		}
	})

	rec := h.do(request{method: "GET", path: "/v1/portfolios/trending/daily"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	in := h.f.dailyIn
	if in.PortfolioName != "trending" || !in.NoSnapshot || in.Short || in.NewsLimit != 2 {
		t.Fatalf("use case got %+v: asking must never save a snapshot", in)
	}

	got := decode[map[string]any](t, rec)
	holding := got["holdings"].([]any)[0].(map[string]any)
	if holding["changeRatio"] != 0.1421 || holding["symbol"] != "MRNA" || holding["value"] != 199.67 {
		t.Fatalf("holding = %v", holding)
	}
	if got["aiSummary"] != "all quiet" || got["risk"].(map[string]any)["largestPosition"] != "SNDK" {
		t.Fatalf("body = %v", got)
	}
	if got["targets"].([]any)[0].(map[string]any)["type"] != "stop-loss" || len(got["valuationIssues"].([]any)) != 1 || len(got["attention"].([]any)) != 1 {
		t.Fatalf("body = %v", got)
	}
}

func TestDaily_EmptyReportHasEmptyArraysAndNoRisk(t *testing.T) {
	t.Parallel()

	body := newHarness(t, nil).do(request{method: "GET", path: "/v1/portfolios/p/daily"}).Body.String()

	for _, want := range []string{`"holdings":[]`, `"news":[]`, `"attention":[]`, `"targets":[]`, `"valuationIssues":[]`} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, `"risk"`) || strings.Contains(body, "aiSummary") {
		t.Errorf("risk and aiSummary should be omitted when empty: %s", body)
	}
}

func TestDaily_QueryParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		query      string
		wantStatus int
		wantShort  bool
		wantNews   int
	}{
		{"", http.StatusOK, false, 2},
		{"?short=true", http.StatusOK, true, 2},
		{"?short=false&newsLimit=5", http.StatusOK, false, 5},
		{"?newsLimit=10", http.StatusOK, false, 10},
		{"?short=maybe", http.StatusUnprocessableEntity, false, 0},
		{"?newsLimit=0", http.StatusUnprocessableEntity, false, 0},
		{"?newsLimit=11", http.StatusUnprocessableEntity, false, 0},
		{"?newsLimit=lots", http.StatusUnprocessableEntity, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t, nil)
			rec := h.do(request{method: "GET", path: "/v1/portfolios/p/daily" + tt.query})

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				if h.f.dailyRun {
					t.Fatal("a bad query must not reach the use case")
				}
				return
			}
			if h.f.dailyIn.Short != tt.wantShort || h.f.dailyIn.NewsLimit != tt.wantNews {
				t.Fatalf("use case got %+v", h.f.dailyIn)
			}
		})
	}
}

func TestUseCaseErrorsBecomeTheRightResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"validation", domain.Invalid("quantity must be greater than 0"), http.StatusUnprocessableEntity, "invalid"},
		{"selling more than held", fmt.Errorf("sell MRNA: %w", domain.ErrInsufficientQuantity), http.StatusUnprocessableEntity, "invalid"},
		{"portfolio not found", fmt.Errorf("%w: %q", domain.ErrPortfolioNotFound, "nope"), http.StatusNotFound, "not_found"},
		{"already exists", domain.ErrPortfolioAlreadyExists, http.StatusConflict, "conflict"},
		{"data changed", domain.ErrConcurrentChange, http.StatusConflict, "conflict"},
		{"provider down", fmt.Errorf("%w: fx rate: %w", domain.ErrUpstream, errors.New("dial tcp: timeout")), http.StatusBadGateway, "unavailable"},
		{"anything else", errors.New("disk on fire"), http.StatusInternalServerError, "internal"},
	}

	endpoints := []request{
		{method: "GET", path: "/v1/portfolios"},
		{method: "POST", path: "/v1/portfolios", body: `{"name":"p","baseCurrency":"EUR"}`},
		{method: "POST", path: "/v1/portfolios/p/trades", body: `{"trades":[]}`},
		{method: "GET", path: "/v1/portfolios/p/summary"},
		{method: "GET", path: "/v1/portfolios/p/daily"},
	}

	for _, tt := range tests {
		for _, ep := range endpoints {
			t.Run(tt.name+" "+ep.method+" "+ep.path, func(t *testing.T) {
				t.Parallel()

				h := newHarness(t, func(f *fakes) {
					f.listErr, f.createErr, f.tradesErr, f.summaryErr, f.dailyErr = tt.err, tt.err, tt.err, tt.err, tt.err
				})

				rec := h.do(ep)

				if rec.Code != tt.wantStatus {
					t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body.String())
				}
				if got := decode[errResp](t, rec).Error.Code; got != tt.wantCode {
					t.Fatalf("code = %q, want %q", got, tt.wantCode)
				}
			})
		}
	}
}

func TestErrorMessages_NeverLeakInternalsOrSecrets(t *testing.T) {
	t.Parallel()

	leaky := fmt.Errorf("%w: quote: %w", domain.ErrUpstream,
		errors.New(`Get "https://finnhub.io/api/v1/quote?symbol=MRNA&token=SUPERSECRETKEY": dial tcp: lookup finnhub.io: no such host`))
	internal := errors.New(`SQL logic error: no such column: provider_symbol at /Users/me/.tick/tick.db`)

	tests := []struct {
		name string
		err  error
		code int
	}{
		{"provider failure with a key in the URL", leaky, http.StatusBadGateway},
		{"internal failure with paths and SQL", internal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t, func(f *fakes) { f.summaryErr = tt.err })
			rec := h.do(request{method: "GET", path: "/v1/portfolios/p/summary"})

			if rec.Code != tt.code {
				t.Fatalf("status = %d", rec.Code)
			}

			body := rec.Body.String()
			for _, secret := range []string{"SUPERSECRETKEY", "finnhub.io", "provider_symbol", "/Users/me", "dial tcp"} {
				if strings.Contains(body, secret) {
					t.Errorf("response leaks %q: %s", secret, body)
				}
			}

			logs := h.logs.String()
			if strings.Contains(logs, "SUPERSECRETKEY") {
				t.Errorf("the log leaks the API key: %s", logs)
			}
			if !strings.Contains(logs, "REDACTED") && tt.code == http.StatusBadGateway {
				t.Errorf("the log should still record the failure, with the key hidden: %s", logs)
			}
		})
	}
}

func TestMessagesForCallerMistakesAreShown(t *testing.T) {
	t.Parallel()

	h := newHarness(t, func(f *fakes) { f.summaryErr = fmt.Errorf("%w: %q", domain.ErrPortfolioNotFound, "nope") })
	rec := h.do(request{method: "GET", path: "/v1/portfolios/nope/summary"})

	if got := decode[errResp](t, rec).Error.Message; !strings.Contains(got, `portfolio not found: "nope"`) {
		t.Fatalf("message = %q: the caller needs to be told what was wrong", got)
	}
}

func TestPanicsBecomeA500AndAreLogged(t *testing.T) {
	t.Parallel()

	h := newHarness(t, func(f *fakes) { f.panicOnList = true })
	rec := h.do(request{method: "GET", path: "/v1/portfolios"})

	if rec.Code != http.StatusInternalServerError || decode[errResp](t, rec).Error.Code != "internal" {
		t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "kaboom") {
		t.Fatal("the panic value must not reach the caller")
	}
	if !strings.Contains(h.logs.String(), "kaboom") {
		t.Fatal("the panic should be logged")
	}
}

func TestRequestLogNeverContainsTheTokenOrTheBody(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	h.do(request{method: "POST", path: "/v1/portfolios", body: `{"name":"private-portfolio-name","baseCurrency":"EUR"}`})

	logs := h.logs.String()
	if !strings.Contains(logs, "method=POST") || !strings.Contains(logs, "status=201") || !strings.Contains(logs, "/v1/portfolios") {
		t.Fatalf("expected a request line, got: %s", logs)
	}
	for _, secret := range []string{token, "private-portfolio-name", "Bearer"} {
		if strings.Contains(logs, secret) {
			t.Errorf("log contains %q: %s", secret, logs)
		}
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	rec := newHarness(t, nil).do(request{method: "GET", path: "/healthz", noAuth: true})

	if got := decode[map[string]string](t, rec); rec.Code != http.StatusOK || got["status"] != "ok" {
		t.Fatalf("status %d body %v", rec.Code, got)
	}
}

func TestCreatePortfolio_RejectsABadBody(t *testing.T) {
	t.Parallel()

	h := newHarness(t, nil)
	rec := h.do(request{method: "POST", path: "/v1/portfolios", body: `{"name": 5}`})

	if rec.Code != http.StatusBadRequest || h.f.createIn.PortfolioName != "" {
		t.Fatalf("status = %d, use case got %+v", rec.Code, h.f.createIn)
	}
}

func TestNew_WorksWithoutALogger(t *testing.T) {
	t.Parallel()

	f := newFakes()
	h := api.New(api.Deps{ListPortfolios: f.lister()}, api.Options{Token: token})

	req := httptest.NewRequest("GET", "/v1/portfolios", nil)
	req.Host = "localhost"
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
