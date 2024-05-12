package aggtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"

	aggendpoint "github.com/bzawada1/location-app-obu-service/go-kit-example/aggsvc/agg_endpoint"
	"github.com/bzawada1/location-app-obu-service/go-kit-example/aggsvc/aggservice"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPClient(instance string, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (aggservice.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}

	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
	var options []httptransport.ClientOption

	if zipkinTracer != nil {
		// Zipkin HTTP Client Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing client can be instantiated
		// without an operation name and fed to each Go kit endpoint as ClientOption.
		// In the latter case, the operation name will be the endpoint's http method.
		options = append(options, zipkin.HTTPClientTrace(zipkinTracer))
	}

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var aggEndpoint endpoint.Endpoint
	{
		aggEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/aggregate"),
			encodeHTTPGenericRequest,
			decodeHTTPAggregateResponse,
			append(options, httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)))...,
		).Endpoint()
		aggEndpoint = opentracing.TraceClient(otTracer, "Sum")(aggEndpoint)
		if zipkinTracer != nil {
			aggEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Sum")(aggEndpoint)
		}
		aggEndpoint = limiter(aggEndpoint)
		aggEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Aggregate",
			Timeout: 30 * time.Second,
		}))(aggEndpoint)
	}

	// The Concat endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var calculateEndpoint endpoint.Endpoint
	{
		calculateEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/invoice"),
			encodeHTTPGenericRequest,
			decodeHTTPCalculateResponse,
			append(options, httptransport.ClientBefore(opentracing.ContextToHTTP(otTracer, logger)))...,
		).Endpoint()
		calculateEndpoint = opentracing.TraceClient(otTracer, "Concat")(calculateEndpoint)
		if zipkinTracer != nil {
			calculateEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Concat")(calculateEndpoint)
		}
		calculateEndpoint = limiter(calculateEndpoint)
		calculateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Calculate",
			Timeout: 30 * time.Second,
		}))(calculateEndpoint)
	}

	return aggendpoint.Set{
		AggregateEndpoint: aggEndpoint,
		CalculateEndpoint: calculateEndpoint,
	}, nil
}

func NewHTTPHandler(endpoints aggendpoint.Set, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	m := http.NewServeMux()
	m.Handle("/aggregate", httptransport.NewServer(
		endpoints.AggregateEndpoint,
		decodeHTTPAggregateRequest,
		encodeHTTPGenericResponse,
		options...,
	))
	// m.Handle("/invoice", httptransport.NewServer(
	// 	endpoints.CalculateEndpoint,
	// 	decodeHTTPCalculateRequest,
	// 	encodeHTTPGenericResponse,
	// 	options...,
	// ))
	return m
}

func errorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	fmt.Println("this is coming from the error encoder", err)
}

func decodeHTTPAggregateResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	resp := aggendpoint.AggregateResponse{}
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPCalculateResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	resp := aggendpoint.CalculateResponse{}
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeHTTPAggregateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := aggendpoint.AggregateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPCalculateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := aggendpoint.CalculateRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = io.NopCloser(&buf)
	return nil
}
