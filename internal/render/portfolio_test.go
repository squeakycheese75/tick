package render_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func summaryReport() domain.PortfoloSummaryReport {
	return domain.PortfoloSummaryReport{
		PortfolioName: "trending",
		BaseCurrency:  "EUR",
		TotalValue:    1500,
		TotalCost:     1400,
		TotalPnL:      100,
		TotalPnLPct:   100.0 / 1400.0,
		TotalAssetPnL: 80,
		TotalFXPnL:    -30,
		Positions: []domain.SummaryPosition{
			{Symbol: "AAPL", Quantity: 10, QuotedPrice: 110, InstrumentCurrency: "USD", MarketValueBase: 880, CostBasisBase: 900, UnrealizedPnL: -20, AssetPnL: 80, FXPnL: -100, FXKnown: true, UnrealizedPnLPct: -0.0222},
			{Symbol: "FLAT", Quantity: 1, QuotedPrice: 5, InstrumentCurrency: "USD", MarketValueBase: 5, CostBasisBase: 4, UnrealizedPnL: 1, AssetPnL: 1, FXPnL: 0.001, FXKnown: true, UnrealizedPnLPct: 0.25},
			{Symbol: "OLD", Quantity: 2, QuotedPrice: 7, InstrumentCurrency: "USD", MarketValueBase: 14, CostBasisBase: 10, UnrealizedPnL: 4, AssetPnL: 4, FXKnown: false, UnrealizedPnLPct: 0.4},
		},
	}
}

func renderSummary(t *testing.T, mutate func(*render.PortfolioSummaryOptions)) string {
	t.Helper()

	opts := render.DefaultPortfolioSummaryOptions()
	opts.Color = false
	if mutate != nil {
		mutate(&opts)
	}

	var buf bytes.Buffer
	if err := render.RenderPortfolioSummary(&buf, summaryReport(), opts); err != nil {
		t.Fatalf("render: %v", err)
	}

	return buf.String()
}

func TestRenderPortfolioSummary_SplitsAssetAndFX(t *testing.T) {
	t.Parallel()

	got := renderSummary(t, nil)
	lines := strings.Split(got, "\n")

	row := func(symbol string) string {
		for _, l := range lines {
			if strings.HasPrefix(l, symbol) {
				return l
			}
		}
		t.Fatalf("no row for %s in:\n%s", symbol, got)
		return ""
	}

	for _, want := range []string{
		"trending · EUR",
		"from assets",
		"+80.00 EUR",
		"-30.00 EUR",
		"(FX split covers 2 of 3 positions",
		"Positions\n",
		"ASSET",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}

	if r := row("AAPL"); !strings.Contains(r, "+80.00") || !strings.Contains(r, "-100.00") || !strings.Contains(r, "-2.22%") {
		t.Errorf("AAPL row should show asset and FX parts: %q", r)
	}
	if r := row("FLAT"); !strings.Contains(r, "·") {
		t.Errorf("FX too small to show should be a dot: %q", r)
	}
	if r := row("OLD"); strings.Count(r, " -") < 2 {
		t.Errorf("a position with no purchase rate should show dashes for asset and FX: %q", r)
	}
	if strings.Contains(got, "Base currency") {
		t.Errorf("the base currency is in the title, not a separate line:\n%s", got)
	}
	if strings.Contains(got, "PRICE") || strings.Contains(got, "USD") {
		t.Errorf("prices should be hidden by default:\n%s", got)
	}
}

func TestRenderPortfolioSummary_ShowsPricesOnRequest(t *testing.T) {
	t.Parallel()

	got := renderSummary(t, func(o *render.PortfolioSummaryOptions) { o.ShowPrices = true })

	if !strings.Contains(got, "PRICE") || !strings.Contains(got, "110.00 USD") {
		t.Fatalf("expected prices:\n%s", got)
	}
}

func TestRenderPortfolioSummary_NoSplitLinesWithoutTrackedPositions(t *testing.T) {
	t.Parallel()

	r := summaryReport()
	for i := range r.Positions {
		r.Positions[i].FXKnown = false
	}

	var buf bytes.Buffer
	opts := render.DefaultPortfolioSummaryOptions()
	opts.Color = false
	if err := render.RenderPortfolioSummary(&buf, r, opts); err != nil {
		t.Fatalf("render: %v", err)
	}

	if strings.Contains(buf.String(), "from FX") {
		t.Fatalf("no FX split expected:\n%s", buf.String())
	}
}

func TestRenderPortfolioSummary_ColorsAndEmpty(t *testing.T) {
	t.Parallel()

	got := renderSummary(t, func(o *render.PortfolioSummaryOptions) { o.Color = true })
	if !strings.Contains(got, "\033[31m") || !strings.Contains(got, "\033[32m") {
		t.Fatalf("expected red and green:\n%q", got)
	}

	var buf bytes.Buffer
	empty := domain.PortfoloSummaryReport{PortfolioName: "p", BaseCurrency: "EUR"}
	if err := render.RenderPortfolioSummary(&buf, empty, render.DefaultPortfolioSummaryOptions()); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "No positions") {
		t.Fatalf("expected empty message:\n%s", buf.String())
	}
}

func TestRenderPortfolioSummary_TotalsLineUp(t *testing.T) {
	t.Parallel()

	got := renderSummary(t, nil)

	// Every totals line ends its amount at the same column: that is where "EUR" ends.
	end := -1
	for _, l := range strings.Split(got, "\n") {
		isTotals := strings.HasPrefix(l, "Total value") || strings.HasPrefix(l, "Total cost") ||
			strings.HasPrefix(l, "Total PnL") || strings.HasPrefix(l, "  from ")
		if !isTotals {
			continue
		}

		i := strings.Index(l, "EUR") + len("EUR")
		if end == -1 {
			end = i
		}
		if i != end {
			t.Errorf("amount ends at column %d, want %d: %q", i, end, l)
		}
	}
	if end == -1 {
		t.Fatalf("no totals lines found:\n%s", got)
	}

	if !strings.Contains(got, "EUR  +7.14%") {
		t.Errorf("percentage should sit in its own column after the PnL amount:\n%s", got)
	}
}

func TestRenderPortfolioSummary_SaysWhenPricesAreFromAnEarlierDay(t *testing.T) {
	t.Parallel()

	friday := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	monday := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		asOf   time.Time
		now    time.Time
		wanted string
	}{
		{"before the next session opens", friday, monday, "Prices as of Fri 9 Oct"},
		{"same day", monday.Add(-time.Hour), monday, ""},
		{"provider gave no time", time.Time{}, monday, ""},
		{"caller gave no clock", friday, time.Time{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := summaryReport()
			r.PricesAsOf = tt.asOf

			var buf bytes.Buffer
			opts := render.DefaultPortfolioSummaryOptions()
			opts.Color = false
			opts.Now = tt.now
			if err := render.RenderPortfolioSummary(&buf, r, opts); err != nil {
				t.Fatalf("render: %v", err)
			}

			got := buf.String()
			if tt.wanted == "" && strings.Contains(got, "Prices as of") {
				t.Fatalf("unexpected note:\n%s", got)
			}
			if tt.wanted != "" && !strings.Contains(got, tt.wanted+"\n") {
				t.Fatalf("missing %q:\n%s", tt.wanted, got)
			}
		})
	}
}

func TestPortfolioList(t *testing.T) {
	t.Parallel()

	out := domain.ListPortfoliosOutput{Portfolios: []domain.PortfolioListItem{
		{Name: "main", BaseCurrency: "EUR", PositionCount: 0},
		{Name: "trending", BaseCurrency: "EUR", PositionCount: 5},
	}}

	var buf bytes.Buffer
	if err := render.PortfolioList(&buf, out, "trending"); err != nil {
		t.Fatalf("render: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "  main") || !strings.HasPrefix(lines[2], "* trending") {
		t.Fatalf("the default portfolio should be starred:\n%s", buf.String())
	}
	if !strings.HasSuffix(lines[2], "5") {
		t.Fatalf("expected the position count at the end:\n%s", buf.String())
	}

	buf.Reset()
	if err := render.PortfolioList(&buf, domain.ListPortfoliosOutput{}, "main"); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "tick portfolio create") {
		t.Fatalf("an empty list should say how to create one:\n%s", buf.String())
	}
}

func TestDefaultPortfolioSet(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := render.DefaultPortfolioSet(&buf, "trending"); err != nil {
		t.Fatalf("render: %v", err)
	}
	if buf.String() != "Default portfolio: trending\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestDailyReport_AlignsQuantitiesAndTheShortReportDoesNotDoubleSpace(t *testing.T) {
	t.Parallel()

	holdings := domain.HoldingSummary{Holdings: []domain.Holding{
		{Symbol: "AAA", Quantity: 0.143884, MarketValueBase: 10, QuotedPrice: 1, PriceCurrency: "USD"},
		{Symbol: "BBB", Quantity: 3.05064, MarketValueBase: 10, QuotedPrice: 1, PriceCurrency: "USD"},
		{Symbol: "CCC", Quantity: 0.995024, MarketValueBase: 10, QuotedPrice: 1, PriceCurrency: "USD"},
	}}

	var daily bytes.Buffer
	err := render.DailyReport(&daily, domain.GetDailyReportOutput{DailyReport: domain.DailyReport{
		Portfolio: domain.PortfolioSummary{Name: "p", BaseCurrency: "EUR"}, TopHoldings: holdings,
	}}, render.DefaultDailyReportOptions())
	if err != nil {
		t.Fatalf("render daily: %v", err)
	}

	// With one shared number of decimals the decimal point sits in the same column on every row.
	dots := map[int]bool{}
	for _, l := range strings.Split(daily.String(), "\n") {
		for _, sym := range []string{"AAA", "BBB", "CCC"} {
			if strings.HasPrefix(l, sym) {
				dots[strings.Index(l, ".")] = true
			}
		}
	}
	if len(dots) != 1 {
		t.Fatalf("quantity decimal points are not aligned (columns %v):\n%s", dots, daily.String())
	}
	if !strings.Contains(daily.String(), "3.050640") {
		t.Fatalf("expected one number of decimals for the column:\n%s", daily.String())
	}

	var short bytes.Buffer
	err = render.DailyReport(&short, domain.GetDailyReportOutput{DailyReport: domain.DailyReport{
		Portfolio: domain.PortfolioSummary{Name: "p", BaseCurrency: "EUR"}, TopHoldings: holdings,
	}}, render.ShortDailyReportOptions())
	if err != nil {
		t.Fatalf("render short: %v", err)
	}
	if strings.Contains(short.String(), "\n\n\n") {
		t.Fatalf("the short report should not have two blank lines in a row:\n%q", short.String())
	}
}

func TestDailyReport_ShortShowsOnlyMoversAndNews(t *testing.T) {
	t.Parallel()

	report := domain.DailyReport{
		Portfolio:   domain.PortfolioSummary{Name: "p", BaseCurrency: "EUR", TotalValue: 100},
		TopHoldings: domain.HoldingSummary{Holdings: []domain.Holding{{Symbol: "AAA", Quantity: 1, MarketValueBase: 100, QuotedPrice: 1, PriceCurrency: "USD"}}},
		Risk:        domain.RiskSummary{LargestPosition: "AAA", LargestWeight: 1},
		Attention:   []string{"AAA is 100% of the portfolio"},
		News:        []domain.NewsSummary{{Ticker: "AAA", Headlines: []domain.NewsHeadline{{Title: "AAA does a thing"}}}},
	}

	var short, full bytes.Buffer
	if err := render.DailyReport(&short, domain.GetDailyReportOutput{DailyReport: report}, render.ShortDailyReportOptions()); err != nil {
		t.Fatalf("render short: %v", err)
	}
	if err := render.DailyReport(&full, domain.GetDailyReportOutput{DailyReport: report}, render.DefaultDailyReportOptions()); err != nil {
		t.Fatalf("render full: %v", err)
	}

	for _, want := range []string{"Movers", "News", "AAA does a thing"} {
		if !strings.Contains(short.String(), want) {
			t.Errorf("short report missing %q:\n%s", want, short.String())
		}
	}
	for _, unwanted := range []string{"Risk", "Attention", "Holdings"} {
		if strings.Contains(short.String(), unwanted) {
			t.Errorf("short report should not include %q:\n%s", unwanted, short.String())
		}
	}
	for _, want := range []string{"Holdings", "Risk", "Attention"} {
		if !strings.Contains(full.String(), want) {
			t.Errorf("full report missing %q:\n%s", want, full.String())
		}
	}
}
