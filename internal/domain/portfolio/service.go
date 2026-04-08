package portfolio

// type PositionLister interface {
// 	ListByPortfolio(ctx context.Context, portfolioName string) ([]Position, error)
// }

// type Service struct {
// 	repo PositionLister
// }

// func NewService(repo PositionLister) *Service {
// 	return &Service{repo: repo}
// }

// func (s *Service) ListPositions(ctx context.Context, portfolioName string) ([]Position, error) {
// 	return s.repo.ListByPortfolio(ctx, portfolioName)
// }

// func (s *Service) BuildSummary(portfolioName string, positions []Position) Summary {
// 	views := make([]PositionView, 0, len(positions))
// 	var totalValue float64

// 	for _, position := range positions {
// 		currentPrice := position.AvgCost
// 		marketValue := currentPrice * position.Quantity
// 		totalValue += marketValue

// 		views = append(views, PositionView{
// 			Ticker:       position.Ticker,
// 			Quantity:     position.Quantity,
// 			AvgCost:      position.AvgCost,
// 			Currency:     position.Currency,
// 			CurrentPrice: currentPrice,
// 			MarketValue:  marketValue,
// 		})
// 	}

// 	if totalValue > 0 {
// 		for i := range views {
// 			views[i].Weight = views[i].MarketValue / totalValue
// 		}
// 	}

// 	sort.Slice(views, func(i, j int) bool {
// 		return views[i].MarketValue > views[j].MarketValue
// 	})

// 	return Summary{
// 		PortfolioName: portfolioName,
// 		TotalValue:    totalValue,
// 		Positions:     views,
// 	}
// }
