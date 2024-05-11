package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/bzawada1/location-app-obu-service/types"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	store := NewMemoryStore()
	svc := NewInvoiceAggregator(store)
	svc = NewLogMiddleware(svc)
	svc = NewMetricsMiddleware(svc)
	grpcAddr := os.Getenv("AGG_GRPC_ENDPOINT")
	httpAddr := os.Getenv("AGG_HTTP_ENDPOINT")
	go makeGRPCTransport(grpcAddr, svc)
	makeHTTPTransport(httpAddr, svc)
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	ln, err := net.Listen("TCP", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	server := grpc.NewServer([]grpc.ServerOption{}...)
	types.RegisterAggregatorServer(server, NewAggregatorGRPCServer(svc).server)
	return server.Serve(ln)

}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	aggMetricHandler := newHTTPMetricsHandler("aggregate")
	invMetricHandler := newHTTPMetricsHandler("invoice")
	aggregateHandler := makeHTTPHandlerFunc(aggMetricHandler.instrument(handleAggregate(svc)))
	invoiceHandler := makeHTTPHandlerFunc(invMetricHandler.instrument(handleAggregate(svc)))
	http.HandleFunc("/aggregate", aggregateHandler)
	http.HandleFunc("/invoice", invoiceHandler)
	// http.HandleFunc("/invoice/all", handleGetAllInvoice(svc))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
