package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strconv"

	"github.com/bzawada1/location-app-obu-service/types"
	"google.golang.org/grpc"
)

func main() {
	httpAddr := flag.String("httpAddr", ":3000", "the listen address of the HTTP server")
	grpcAddr := flag.String("grpcAddr", ":3001", "the listen address of the HTTP server")
	flag.Parse()
	store := NewMemoryStore()
	svc := NewInvoiceAggregator(store)
	logSvc := NewLogMiddleware(svc)
	go makeGRPCTransport(*grpcAddr, svc)
	makeHTTPTransport(*httpAddr, logSvc)
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	ln, err := net.Listen("TCP", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	server := grpc.NewServer([]grpc.ServerOption{}...)
	types.RegisterAggregatorServer(server, NewAggregatorGRPCServer(svc))
	return server.Serve(ln)

}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.HandleFunc("/invoice", handleGetInvoice(svc))
	http.ListenAndServe(listenAddr, nil)
}

func handleGetInvoice(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values, ok := r.URL.Query()["obu"]
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing OBU ID"})
			return
		}
		obuID, err := strconv.Atoi(values[0])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid OBU ID"})
			return
		}
		invoice, err := svc.CalculateInvoice(obuID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, invoice)
		return
	}
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		distance := types.Distance{}

		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := svc.AggregateDistance(distance); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
