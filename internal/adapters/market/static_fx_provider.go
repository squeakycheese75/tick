package market

import (
	"context"
	"fmt"
	"strings"
)

type StaticFXProvider struct {
	rates map[string]float64
}

func NewStaticFXProvider() *StaticFXProvider {
	return &StaticFXProvider{
		rates: map[string]float64{
			"EUR:EUR": 1.0,
			"USD:USD": 1.0,
			"GBP:GBP": 1.0,
			"USD:EUR": 0.92,
			"EUR:USD": 1.09,
			"GBP:EUR": 1.17,
			"EUR:GBP": 0.85,
			"USD:GBP": 0.78,
			"GBP:USD": 1.28,
		},
	}
}

func (f *StaticFXProvider) GetRate(_ context.Context, from string, to string) (float64, error) {
	key := strings.ToUpper(from) + ":" + strings.ToUpper(to)
	rate, ok := f.rates[key]
	if !ok {
		return 0, fmt.Errorf("fx rate not found for %s", key)
	}
	return rate, nil
}
