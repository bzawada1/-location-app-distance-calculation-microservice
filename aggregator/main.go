package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"github.com/bzawada1/location-app-obu-service/types"
)

func main() {
	listenAddr := flag.String("listenaddr", ":3000", "the listen address of the HTTP server")
	flag.Parse()
	store := NewMemoryStore()
	svc := NewInvoiceAggregator(store)
	makeHTTPTransport(*listenAddr, svc)
}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.ListenAndServe(listenAddr, nil)
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		distance := types.Distance{}

		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			panic(err)
		}
	}
}
