package main

import (
	// "fmt"
	"net"
	"net/http"
	"os"

	aggendpoint "github.com/bzawada1/location-app-obu-service/go-kit-example/aggsvc/agg_endpoint"
	"github.com/bzawada1/location-app-obu-service/go-kit-example/aggsvc/aggservice"
	"github.com/bzawada1/location-app-obu-service/go-kit-example/aggsvc/aggtransport"
	"github.com/go-kit/log"
)

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	var (
		service     = aggservice.New(logger)
		endpoints   = aggendpoint.New(service, logger)
		httpHandler = aggtransport.NewHTTPHandler(endpoints, logger)
	)

	httpAddr := ":5201"
	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}

	logger.Log("transport", "HTTP", "addr", &httpAddr)
	err = http.Serve(httpListener, httpHandler)
	if err != nil {
		panic(err)
	}
}
