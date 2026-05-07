package domain

import "errors"

var (
	ErrPortfolioNotFound         = errors.New("portfolio not found")
	ErrPositionNotFound          = errors.New("position not found")
	ErrInstrumentNotFound        = errors.New("instrument not found")
	ErrPortfolioAlreadyExists    = errors.New("portfolio already exists")
	ErrPositionAlreadyExists     = errors.New("position already exists")
	ErrInstrumentAlreadyExists   = errors.New("instrument already exists")
	ErrPriceCacheNotFound        = errors.New("price cache not found")
	ErrConsumedPriceNotFound     = errors.New("consumed price not found")
	ErrFXCacheNotFound           = errors.New("fx cache not found")
	ErrPortfolioSnapshotNotFound = errors.New("portfolio snapshot not found")
	ErrTargetNotFound            = errors.New("target not found")
	ErrFXRateNotFound            = errors.New("fx rate not found")
)
