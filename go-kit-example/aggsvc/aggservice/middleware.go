package aggservice

import (
	"context"

	"github.com/bzawada1/location-app-obu-service/types"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	next Service
}

func newLoggingMiddleware() Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			next: next,
		}
	}
}

func (lm loggingMiddleware) Aggregate(_ context.Context, dist types.Distance) error {
	return nil
}

func (lm loggingMiddleware) Calculate(_ context.Context, obuID int) (*types.Invoice, error) {
	return nil, nil
}

type instrumentationMiddleware struct {
	next Service
}

func newInstrumentationMiddleware() Middleware {
	return func(next Service) Service {
		return instrumentationMiddleware{
			next: next,
		}
	}
}

func (lm instrumentationMiddleware) Aggregate(_ context.Context, dist types.Distance) error {
	return nil
}

func (lm instrumentationMiddleware) Calculate(_ context.Context, obuID int) (*types.Invoice, error) {
	return nil, nil
}
