package render

// func TestDailyReport(t *testing.T) {
// 	t.Parallel()

// 	var buf bytes.Buffer

// 	out := domain.GetDailyReportOutput{
// 		DailyReport: domain.DailyReport{
// 			PortfolioName: "main",
// 			BaseCurrency:  "EUR",
// 			TotalValue:    36369.31,
// 			TopHoldings: []domain.TopHoldingReport{
// 				{
// 					Symbol:          "BTC",
// 					Weight:          0.9106,
// 					MarketValueBase: 33118.84,
// 					QuotedPrice:     77815.00,
// 					PriceCurrency:   "USD",
// 					ChangePercent:   0,
// 				},
// 			},
// 			Risk: domain.RiskReport{
// 				LargestPosition:   "BTC",
// 				LargestWeight:     0.9106,
// 				Top3Concentration: 1.0,
// 			},
// 			News: []domain.TickerNewsReport{
// 				{
// 					Ticker: "BTC",
// 					Headlines: []domain.NewsHeadline{
// 						{
// 							Title: "Bitcoin and crypto stocks fall after hearing",
// 							URL:   "https://example.com",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	opts := DefaultDailyReportOptions()
// 	opts.Holdings.Color = false

// 	err := DailyReport(&buf, out, opts)
// 	if err != nil {
// 		t.Fatalf("DailyReport() error = %v", err)
// 	}

// 	got := buf.String()
// 	want := `main  36,369.31 EUR

// Holdings
// BTC     91.06%    33,118.84 EUR  @    77,815.00 USD  → +0.00%

// Risk   Largest: BTC (91.06%)   Top 3: 100.00%   ! High concentration

// News
// BTC:  Bitcoin and crypto stocks fall after hearing
// `

// 	if got != want {
// 		t.Fatalf("unexpected output\n\nwant:\n%s\n\ngot:\n%s", want, got)
// 	}
// }
