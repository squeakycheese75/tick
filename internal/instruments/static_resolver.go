package instruments

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/instruments.json
var instrumentsFS []byte

type ResolvedInstrument struct {
	Symbol         string
	InstrumentType string
	Exchange       string
	QuoteCurrency  string
}

type instrumentJSON struct {
	Symbol         string `json:"symbol"`
	InstrumentType string `json:"asset_type"`
	Exchange       string `json:"exchange"`
	QuoteCurrency  string `json:"quote_currency"`
}

type StaticResolver struct {
	instruments map[string]ResolvedInstrument
}

func NewStaticResolver() (*StaticResolver, error) {
	var raw []instrumentJSON

	if err := json.Unmarshal(instrumentsFS, &raw); err != nil {
		return nil, fmt.Errorf("parse embedded instruments: %w", err)
	}

	m := make(map[string]ResolvedInstrument, len(raw))

	for _, r := range raw {
		symbol := strings.ToUpper(strings.TrimSpace(r.Symbol))

		m[symbol] = ResolvedInstrument{
			Symbol:         symbol,
			InstrumentType: strings.ToLower(r.InstrumentType),
			Exchange:       strings.ToUpper(r.Exchange),
			QuoteCurrency:  strings.ToUpper(r.QuoteCurrency),
		}
	}

	return &StaticResolver{
		instruments: m,
	}, nil
}

func (r *StaticResolver) Resolve(ctx context.Context, symbol string) (ResolvedInstrument, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	if inst, ok := r.instruments[symbol]; ok {
		return inst, nil
	}

	if symbol == "BRKB" {
		return r.instruments["BRK.B"], nil
	}

	return ResolvedInstrument{}, fmt.Errorf(
		"instrument %q not found. try specifying --asset-type, --exchange, and --quote-currency",
		symbol,
	)
}
