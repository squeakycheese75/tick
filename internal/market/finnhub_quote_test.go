package market

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubFinnhub(body string) *FinnhubPriceProvider {
	p := NewFinnhubPriceProvider("key")
	p.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	return p
}

func TestFinnhub_GetQuote_ReportsWhenThePriceLastTraded(t *testing.T) {
	t.Parallel()

	// Friday 2026-10-09 20:00 UTC, the close the previous session's move is measured to.
	p := stubFinnhub(`{"c":225,"d":28,"dp":14.2132,"h":225,"l":197.17,"o":199,"pc":197,"t":1791576000}`)

	q, err := p.GetQuote(context.Background(), GetQuoteParams{Symbol: "MRNA", ProviderSymbol: "MRNA"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	if !q.AsOf.Equal(want) {
		t.Fatalf("AsOf = %v, want %v", q.AsOf, want)
	}
	if q.Price != 225 || q.PreviousClose != 197 {
		t.Fatalf("quote = %+v", q)
	}
}

func TestFinnhub_GetQuote_NoTimestampLeavesAsOfUnset(t *testing.T) {
	t.Parallel()

	q, err := stubFinnhub(`{"c":10,"pc":9}`).GetQuote(context.Background(), GetQuoteParams{ProviderSymbol: "X"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.AsOf.IsZero() {
		t.Fatalf("AsOf = %v, want zero", q.AsOf)
	}
}

func TestFinnhub_GetQuote_NoPriceIsAnError(t *testing.T) {
	t.Parallel()

	if _, err := stubFinnhub(`{"c":0}`).GetQuote(context.Background(), GetQuoteParams{ProviderSymbol: "X"}); err == nil {
		t.Fatal("expected an error when no price is returned")
	}
}
