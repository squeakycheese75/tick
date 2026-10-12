package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

const (
	defaultDailyNewsLimit = 2
	maxDailyNewsLimit     = 10
)

func (s *server) listPortfolios(w http.ResponseWriter, r *http.Request) {
	out, err := s.deps.ListPortfolios.Execute(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}

	items := make([]portfolioListItem, 0, len(out.Portfolios))
	for _, p := range out.Portfolios {
		items = append(items, portfolioListItem{Name: p.Name, BaseCurrency: p.BaseCurrency, PositionCount: p.PositionCount})
	}

	writeJSON(w, http.StatusOK, portfolioListResponse{Portfolios: items})
}

func (s *server) createPortfolio(w http.ResponseWriter, r *http.Request) {
	var req createPortfolioRequest
	if err := decodeJSON(w, r, &req); err != nil {
		s.fail(w, r, err)
		return
	}

	out, err := s.deps.CreatePortfolio.Execute(r.Context(), domain.CreatePortfolioUsecaseInput{
		PortfolioName: strings.TrimSpace(req.Name),
		BaseCurrency:  strings.ToUpper(strings.TrimSpace(req.BaseCurrency)),
	})
	if err != nil {
		s.fail(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, portfolioResponse{Name: out.PortfolioName, BaseCurrency: out.BaseCurrency})
}

func (s *server) importTrades(w http.ResponseWriter, r *http.Request) {
	var req importTradesRequest
	if err := decodeJSON(w, r, &req); err != nil {
		s.fail(w, r, err)
		return
	}

	if len(req.Trades) > maxTradesPerRequest {
		s.fail(w, r, domain.Invalid("at most %d trades per request, got %d", maxTradesPerRequest, len(req.Trades)))
		return
	}

	trades, err := toTradeInputs(req.Trades)
	if err != nil {
		s.fail(w, r, err)
		return
	}

	out, err := s.deps.ImportTrades.Execute(r.Context(), domain.ImportTradesInput{
		PortfolioName: r.PathValue("name"),
		Trades:        trades,
	})
	if err != nil {
		s.fail(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, importTradesResponse{Portfolio: out.PortfolioName, Applied: out.Applied, Skipped: out.Skipped})
}

func (s *server) summary(w http.ResponseWriter, r *http.Request) {
	out, err := s.deps.GetSummary.Execute(r.Context(), domain.GetPortfolioSummaryUsecaseInput{PortfolioName: r.PathValue("name")})
	if err != nil {
		s.fail(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toSummaryResponse(out.Report))
}

// daily never saves a snapshot: asking for a report must not change anything.
func (s *server) daily(w http.ResponseWriter, r *http.Request) {
	short, err := queryBool(r, "short")
	if err != nil {
		s.fail(w, r, err)
		return
	}

	newsLimit, err := queryNewsLimit(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}

	out, err := s.deps.GetDaily.Execute(r.Context(), domain.GetDailyReportInput{
		PortfolioName: r.PathValue("name"),
		NewsLimit:     newsLimit,
		Short:         short,
		NoSnapshot:    true,
	})
	if err != nil {
		s.fail(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toDailyResponse(out))
}

func queryBool(r *http.Request, name string) (bool, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return false, nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, domain.Invalid("%s must be true or false, got %q", name, value)
	}

	return b, nil
}

func queryNewsLimit(r *http.Request) (int, error) {
	value := r.URL.Query().Get("newsLimit")
	if value == "" {
		return defaultDailyNewsLimit, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > maxDailyNewsLimit {
		return 0, domain.Invalid("newsLimit must be a whole number from 1 to %d, got %q", maxDailyNewsLimit, value)
	}

	return n, nil
}
