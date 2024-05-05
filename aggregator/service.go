package main

import (
	"github.com/bzawada1/location-app-obu-service/types"
	"log"
)

type Aggregator interface {
	AggregateDistance(types.Distance) error
}

type InvoiceAggregator struct {
	store Storer
}

type Storer interface {
	Insert(types.Distance) error
}

func NewInvoiceAggregator(store Storer) *InvoiceAggregator {
	return &InvoiceAggregator{
		store: store,
	}
}

func (i *InvoiceAggregator) AggregateDistance(distance types.Distance) error {
	if err := i.store.Insert(distance); err != nil {
		log.Fatal(err)
	}

	return nil
}
