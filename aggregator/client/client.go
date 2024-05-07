package client

import (
	"context"

	"github.com/bzawada1/location-app-obu-service/types"
)

type Client interface {
	Aggregate(context.Context, *types.AggregateRequest) error
	GetInvoice(context.Context, int) (*types.Invoice, error)
}
