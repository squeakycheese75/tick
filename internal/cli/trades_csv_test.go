package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

func TestParseTradesCSV(t *testing.T) {
	t.Parallel()

	full := `Date, Symbol, Side, Quantity, Price, Currency, Fee, FX_Rate, Ref, Exchange, Instrument_Type, Quote_Currency
2026-10-12,MRNA,buy,1.5,200,EUR,1.25,0.9,abc,NASDAQ,equity,USD
`
	minimal := `date,symbol,side,quantity,price,currency
# a comment

2026-10-12,DELL,sell,2,500,USD
12/10/2026,HPE,buy,3,70,USD
2026-10-12T09:30:00Z,SNDK,buy,1,1000,USD
`

	t.Run("reads every column", func(t *testing.T) {
		t.Parallel()

		got, err := parseTradesCSV(strings.NewReader(full))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := domain.TradeInput{
			Ref: "abc", Date: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Symbol: "MRNA",
			Side: "buy", Quantity: 1.5, Price: 200, Currency: "EUR", Fee: 1.25, FXRate: 0.9,
			Exchange: "NASDAQ", InstrumentType: "equity", QuoteCurrency: "USD",
		}
		if len(got) != 1 || got[0] != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("optional columns, comments, blank lines and date layouts", func(t *testing.T) {
		t.Parallel()

		got, err := parseTradesCSV(strings.NewReader(minimal))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 || got[0].Symbol != "DELL" || got[1].Symbol != "HPE" || got[2].Symbol != "SNDK" {
			t.Fatalf("got %+v", got)
		}
		if got[0].Fee != 0 || got[0].Ref != "" {
			t.Fatalf("optional fields should be empty, got %+v", got[0])
		}
	})

	errorCases := []struct {
		name string
		csv  string
		want string
	}{
		{"empty file", "", "file is empty"},
		{"missing column", "date,symbol,side,quantity,price\n", `missing required column "currency"`},
		{"bad date", "date,symbol,side,quantity,price,currency\nyesterday,A,buy,1,1,USD\n", "line 2: invalid date"},
		{"bad number", "date,symbol,side,quantity,price,currency\n2026-10-12,A,buy,1,243.31EUR,USD\n", `line 2: invalid price "243.31EUR"`},
		{"bad fee", "date,symbol,side,quantity,price,currency,fee\n2026-10-12,A,buy,1,1,USD,x\n", "invalid fee"},
		{"short row", "date,symbol,side,quantity,price,currency\n2026-10-12,A,buy\n", "missing quantity"},
		{"missing price", "date,symbol,side,quantity,price,currency\n2026-10-12,A,buy,1,,USD\n", "missing price"},
		{"unterminated quote", "date,symbol,side,quantity,price,currency\n2026-10-12,\"A,buy,1,1,USD\n", "quote"},
		{"bad header", "date,\"symbol\n", "read header"},
	}

	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseTradesCSV(strings.NewReader(tt.csv))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
