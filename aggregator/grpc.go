package main

import "github.com/bzawada1/location-app-obu-service/types"

type GRPCAggregatorServer struct {
	server types.UnimplementedAggregatorServer
	svc    Aggregator
}

func NewAggregatorGRPCServer(svc Aggregator) GRPCAggregatorServer {
	return GRPCAggregatorServer{
		svc: svc,
	}
}

func (s *GRPCAggregatorServer) AggregateDistance(req types.AggregateRequest) error {
	distance := types.Distance{
		OBUID: int(req.ObuID),
		Value: req.Value,
		Unix:  req.Unix,
	}
	return s.svc.AggregateDistance(distance)
}
