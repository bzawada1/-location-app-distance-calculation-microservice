package main

import (
	"context"

	"github.com/bzawada1/location-app-obu-service/types"
)

type GRPCAggregatorServer struct {
	server types.UnimplementedAggregatorServer
	svc    Aggregator
}

func NewAggregatorGRPCServer(svc Aggregator) GRPCAggregatorServer {
	return GRPCAggregatorServer{
		svc: svc,
	}
}

func (s *GRPCAggregatorServer) Aggregate(ctx context.Context, req *types.AggregateRequest) (*types.None, error) {
	distance := types.Distance{
		OBUID: int(req.ObuID),
		Value: req.Value,
		Unix:  req.Unix,
	}
	return &types.None{}, s.svc.AggregateDistance(distance)
}
