package domain

import "errors"

var (
	ErrPortfolioNotFound  = errors.New("portfolio not found")
	ErrPositionNotFound   = errors.New("position not found")
	ErrInstrumentNotFound = errors.New("instrument not found")
)
