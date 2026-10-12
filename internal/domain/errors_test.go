package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
)

func TestKindOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want domain.ErrorKind
	}{
		{"nothing is not an error", nil, domain.KindInternal},
		{"unknown error", errors.New("disk on fire"), domain.KindInternal},
		{"validation failure", domain.Invalid("symbol is required"), domain.KindInvalid},
		{"selling more than is held", domain.ErrInsufficientQuantity, domain.KindInvalid},
		{"portfolio not found", domain.ErrPortfolioNotFound, domain.KindNotFound},
		{"position not found", domain.ErrPositionNotFound, domain.KindNotFound},
		{"instrument not found", domain.ErrInstrumentNotFound, domain.KindNotFound},
		{"target not found", domain.ErrTargetNotFound, domain.KindNotFound},
		{"portfolio exists", domain.ErrPortfolioAlreadyExists, domain.KindConflict},
		{"position exists", domain.ErrPositionAlreadyExists, domain.KindConflict},
		{"instrument exists", domain.ErrInstrumentAlreadyExists, domain.KindConflict},
		{"data changed under the request", domain.ErrConcurrentChange, domain.KindConflict},
		{"provider failed", domain.ErrUpstream, domain.KindUnavailable},
		{"internal cache misses are not the caller's concern", domain.ErrPriceCacheNotFound, domain.KindInternal},
		{"wrapped once", fmt.Errorf("get portfolio: %w", domain.ErrPortfolioNotFound), domain.KindNotFound},
		{"wrapped twice", fmt.Errorf("a: %w", fmt.Errorf("b: %w", domain.Invalid("bad"))), domain.KindInvalid},
		{"a cause that is also upstream", fmt.Errorf("%w: fx rate: %w", domain.ErrUpstream, errors.New("timeout")), domain.KindUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := domain.KindOf(tt.err); got != tt.want {
				t.Fatalf("KindOf(%v) = %s, want %s", tt.err, got, tt.want)
			}
		})
	}
}

func TestInvalid_ReadsAsItsMessageAndMatchesTheSentinel(t *testing.T) {
	t.Parallel()

	err := domain.Invalid("side must be buy or sell, got %q", "hold")

	if err.Error() != `side must be buy or sell, got "hold"` {
		t.Fatalf("message = %q", err.Error())
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal("a validation failure should match ErrInvalidInput")
	}

	var typed *domain.InvalidInputError
	if !errors.As(fmt.Errorf("trades[3]: %w", err), &typed) || typed.Message == "" {
		t.Fatal("the typed error should be reachable through wrapping")
	}
	if errors.Is(errors.New("other"), domain.ErrInvalidInput) {
		t.Fatal("an unrelated error must not match")
	}
}

func TestErrorKind_String(t *testing.T) {
	t.Parallel()

	want := map[domain.ErrorKind]string{
		domain.KindInternal:    "internal",
		domain.KindInvalid:     "invalid",
		domain.KindNotFound:    "not_found",
		domain.KindConflict:    "conflict",
		domain.KindUnavailable: "unavailable",
	}
	for kind, name := range want {
		if kind.String() != name {
			t.Errorf("%d = %q, want %q", int(kind), kind.String(), name)
		}
	}
}

func TestEveryValidationFailureIsInvalidInput(t *testing.T) {
	t.Parallel()

	bad := map[string]error{
		"empty trades input":     domain.ImportTradesInput{}.Validate(),
		"empty portfolio import": domain.ImportPortfolioInput{}.Validate(),
		"empty position":         domain.ImportPositionInput{}.Validate(),
		"empty create":           (&domain.CreatePortfolioUsecaseInput{}).Validate(),
		"empty trade":            domain.TradeInput{}.Validate(),
	}

	for name, err := range bad {
		if err == nil {
			t.Errorf("%s: expected a validation error", name)
			continue
		}
		if domain.KindOf(err) != domain.KindInvalid {
			t.Errorf("%s: %v is %s, want invalid", name, err, domain.KindOf(err))
		}
	}
}
