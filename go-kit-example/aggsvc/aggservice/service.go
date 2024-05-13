package aggservice

import (
	"context"
	// "os"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/go-kit/log"
)

const basePrice = 3.15

type Service interface {
	Aggregate(context.Context, types.Distance) error
	Calculate(context.Context, int) (*types.Invoice, error)
}
type Storer interface {
	Insert(types.Distance) error
	Get(int) (float64, error)
}

type BasicService struct {
	store Storer
}

func newBasicService(store Storer) Service {
	return &BasicService{
		store: store,
	}
}

func (s *BasicService) Aggregate(ctx context.Context, dist types.Distance) error {
	return s.store.Insert(dist)
}

func (s *BasicService) Calculate(_ context.Context, obuID int) (*types.Invoice, error) {
	dist, err := s.store.Get(obuID)
	if err != nil {
		return nil, err
	}
	inv := &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		TotalAmount:   basePrice * dist,
	}
	return inv, nil
}

func New(logger log.Logger) Service {
	// // var logger log.Logger
	// logger = log.NewLogfmtLogger(os.Stdout)
	// logger = log.With(logger, "service", "aggregator")
	var svc Service
	{
		svc = newBasicService(NewMemoryStore())
		svc = newLoggingMiddleware(logger)(svc)
		svc = newInstrumentationMiddleware()(svc)
	}

	return svc
}
