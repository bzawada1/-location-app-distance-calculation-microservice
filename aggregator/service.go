package main

import (
	"fmt"
	"log"

	"github.com/bzawada1/location-app-obu-service/types"
)

const basePrice = 0.15

type Aggregator interface {
	AggregateDistance(types.Distance) error
	CalculateInvoice(int) (*types.Invoice, error)
}

type InvoiceAggregator struct {
	store Storer
}

type Storer interface {
	Insert(types.Distance) error
	Get(int) (float64, error)
}

func NewInvoiceAggregator(store Storer) *InvoiceAggregator {
	return &InvoiceAggregator{
		store: store,
	}
}

func (i *InvoiceAggregator) AggregateDistance(distance types.Distance) error {
	fmt.Println("processing and inserting distance in the storage %s", distance)
	if err := i.store.Insert(distance); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (i *InvoiceAggregator) DistanceSum(obuID int) (float64, error) {
	return i.store.Get(obuID)
}

func (i *InvoiceAggregator) CalculateInvoice(obuID int) (*types.Invoice, error) {
	dist, err := i.store.Get(obuID)
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
