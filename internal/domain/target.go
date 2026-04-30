package domain

import (
	"time"
)

type TargetType string

const (
	TargetTypeTakeProfit TargetType = "take-profit"
	TargetTypeStopLoss   TargetType = "stop-loss"
)

type Target struct {
	ID            int64
	PortfolioID   int64
	Symbol        string
	Type          TargetType
	TargetPrice   float64
	QuoteCurrency string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
