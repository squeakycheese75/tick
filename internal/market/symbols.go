package market

import "github.com/squeakycheese75/tick/internal/domain"

var DefaultSymbols = map[string][]domain.ProviderSymbol{
	"GOLD": {
		{Provider: "yahoo", Symbol: "GC=F"},
	},
	"MSTR": {
		{Provider: "finnhub", Symbol: "MSTR"},
		{Provider: "yahoo", Symbol: "MSTR"},
	},
	"SILVER": {
		{Provider: "yahoo", Symbol: "SI=F"},
	},
}
