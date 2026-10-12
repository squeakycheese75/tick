package cli

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

var requiredTradeColumns = []string{"date", "symbol", "side", "quantity", "price", "currency"}

var tradeDateLayouts = []string{"2006-01-02", time.RFC3339, "02/01/2006"}

// parseTradesCSV reads trades from a CSV with a header row. Required columns are
// date, symbol, side, quantity, price and currency. Optional columns are fee,
// fx_rate, ref, exchange, instrument_type and quote_currency. Blank lines and
// lines starting with # are ignored. Errors name the offending line.
func parseTradesCSV(r io.Reader) ([]domain.TradeInput, error) {
	reader := csv.NewReader(r)
	reader.Comment = '#'
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("file is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	columns, err := indexColumns(header)
	if err != nil {
		return nil, err
	}

	var trades []domain.TradeInput

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return trades, nil
		}
		if err != nil {
			return nil, err
		}

		line, _ := reader.FieldPos(0)

		trade, err := parseTradeRecord(columns, record)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		trades = append(trades, trade)
	}
}

func indexColumns(header []string) (map[string]int, error) {
	columns := make(map[string]int, len(header))
	for i, name := range header {
		columns[strings.ToLower(strings.TrimSpace(name))] = i
	}

	for _, name := range requiredTradeColumns {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("missing required column %q", name)
		}
	}

	return columns, nil
}

func parseTradeRecord(columns map[string]int, record []string) (domain.TradeInput, error) {
	get := func(name string) string {
		i, ok := columns[name]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	date, err := parseTradeDate(get("date"))
	if err != nil {
		return domain.TradeInput{}, err
	}

	numbers := map[string]float64{}
	for _, name := range []string{"quantity", "price", "fee", "fx_rate"} {
		v, err := parseNumber(name, get(name), isRequiredTradeNumber(name))
		if err != nil {
			return domain.TradeInput{}, err
		}
		numbers[name] = v
	}

	return domain.TradeInput{
		Ref:            get("ref"),
		Date:           date,
		Symbol:         get("symbol"),
		Side:           domain.TradeSide(get("side")),
		Quantity:       numbers["quantity"],
		Price:          numbers["price"],
		Currency:       get("currency"),
		Fee:            numbers["fee"],
		FXRate:         numbers["fx_rate"],
		Exchange:       get("exchange"),
		InstrumentType: get("instrument_type"),
		QuoteCurrency:  get("quote_currency"),
	}, nil
}

func parseTradeDate(value string) (time.Time, error) {
	for _, layout := range tradeDateLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", value)
}

func isRequiredTradeNumber(name string) bool {
	return name == "quantity" || name == "price"
}

// parseNumber reads a number. An empty cell is zero for optional columns and an error for required ones.
func parseNumber(name, value string, required bool) (float64, error) {
	if value == "" {
		if required {
			return 0, fmt.Errorf("missing %s", name)
		}
		return 0, nil
	}

	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}

	return n, nil
}
