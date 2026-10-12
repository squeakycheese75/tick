package domain

import (
	"errors"
	"fmt"
)

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

var ErrInsufficientQuantity = errors.New("insufficient quantity")

// Errors a caller can act on. Delivery layers map them to their own responses with KindOf.
var (
	// ErrInvalidInput marks a request that cannot be acted on as given. Use Invalid to make one.
	ErrInvalidInput = errors.New("invalid input")
	// ErrConcurrentChange means the data changed while a request was being handled, so it
	// can be run again.
	ErrConcurrentChange = errors.New("data changed while the request was handled")
	// ErrUpstream means a price, rate or news provider failed or could not be reached.
	ErrUpstream = errors.New("provider unavailable")
)

// InvalidInputError is a validation failure. It reads as its message and matches ErrInvalidInput.
type InvalidInputError struct {
	Message string
}

func (e *InvalidInputError) Error() string { return e.Message }

// Is lets errors.Is(err, ErrInvalidInput) find any validation failure.
func (e *InvalidInputError) Is(target error) bool { return target == ErrInvalidInput }

// Invalid builds a validation failure from a formatted message.
func Invalid(format string, args ...any) error {
	return &InvalidInputError{Message: fmt.Sprintf(format, args...)}
}

// ErrorKind is what sort of problem an error is, independent of how it is delivered.
type ErrorKind int

// Kinds of error, from the point of view of whoever made the request.
const (
	// KindInternal is anything not recognised: a bug or an unexpected failure.
	KindInternal ErrorKind = iota
	// KindInvalid: the request is wrong as given, or cannot apply to the current data.
	KindInvalid
	// KindNotFound: something the request names does not exist.
	KindNotFound
	// KindConflict: the request clashes with existing data, or the data changed underneath it.
	KindConflict
	// KindUnavailable: a provider the request depends on failed.
	KindUnavailable
)

func (k ErrorKind) String() string {
	switch k {
	case KindInvalid:
		return "invalid"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindUnavailable:
		return "unavailable"
	default:
		return "internal"
	}
}

// KindOf classifies err. Anything it does not recognise is KindInternal.
func KindOf(err error) ErrorKind {
	switch {
	case err == nil:
		return KindInternal
	case isAny(err, ErrInvalidInput, ErrInsufficientQuantity):
		return KindInvalid
	case isAny(err, ErrPortfolioNotFound, ErrPositionNotFound, ErrInstrumentNotFound, ErrTargetNotFound):
		return KindNotFound
	case isAny(err, ErrPortfolioAlreadyExists, ErrPositionAlreadyExists, ErrInstrumentAlreadyExists, ErrConcurrentChange):
		return KindConflict
	case isAny(err, ErrUpstream):
		return KindUnavailable
	default:
		return KindInternal
	}
}

func isAny(err error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(err, t) {
			return true
		}
	}

	return false
}
