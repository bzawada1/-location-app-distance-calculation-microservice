package aggservice

import (
	"context"
	"time"

	"github.com/go-kit/log"

	"github.com/bzawada1/location-app-obu-service/types"
)

type Middleware func(Service) Service

type loggingMiddleware struct {
	next Service
	log  log.Logger
}

func newLoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			next: next,
			log:  logger,
		}
	}
}

func (lm loggingMiddleware) Aggregate(ctx context.Context, dist types.Distance) (err error) {
	defer func(start time.Time) {
		lm.log.Log("took", time.Since(start), "obu", dist.OBUID, "distance", dist.Value, "err", err)
	}(time.Now())
	return lm.next.Aggregate(ctx, dist)
}

func (lm loggingMiddleware) Calculate(ctx context.Context, obuID int) (*types.Invoice, error) {
	return lm.next.Calculate(ctx, obuID)
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

func (im instrumentationMiddleware) Aggregate(ctx context.Context, dist types.Distance) error {
	return im.next.Aggregate(ctx, dist)
}

func (im instrumentationMiddleware) Calculate(ctx context.Context, obuID int) (*types.Invoice, error) {
	return im.next.Calculate(ctx, obuID)
}
