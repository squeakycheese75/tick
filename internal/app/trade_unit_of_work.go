package app

import (
	"context"

	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

// tradeUnitOfWork presents the repository's transaction support as the use case's interface.
// The repository returns concrete types, which Go will not convert to an interface on its own.
type tradeUnitOfWork struct {
	uow *repository.TradeUnitOfWork
}

func newTradeUnitOfWork(uow *repository.TradeUnitOfWork) usecase.TradeUnitOfWork {
	return tradeUnitOfWork{uow: uow}
}

func (t tradeUnitOfWork) Reader() usecase.TradeStore {
	return t.uow.Reader()
}

func (t tradeUnitOfWork) Within(ctx context.Context, fn func(usecase.TradeStore) error) error {
	return t.uow.Within(ctx, func(store *repository.TradeStore) error {
		return fn(store)
	})
}
