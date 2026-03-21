package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vektah/gqlparser/v2/ast"
)

var metricsRegisterer = prometheus.WrapRegistererWith(
	prometheus.Labels{"service": "swarm-api"},
	prometheus.DefaultRegisterer,
)

var httpRequestsTotal = promauto.With(metricsRegisterer).NewCounterVec(
	prometheus.CounterOpts{
		Name: "swarm_api_http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "route", "status_code"},
)

var httpRequestDurationSeconds = promauto.With(metricsRegisterer).NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "swarm_api_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
	},
	[]string{"method", "route", "status_code"},
)

var graphqlResolverCallsTotal = promauto.With(metricsRegisterer).NewCounterVec(
	prometheus.CounterOpts{
		Name: "swarm_api_graphql_resolver_calls_total",
		Help: "Total number of GraphQL resolver calls",
	},
	[]string{"operation_type", "resolver_name", "status"},
)

var graphqlResolverDurationSeconds = promauto.With(metricsRegisterer).NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "swarm_api_graphql_resolver_duration_seconds",
		Help:    "GraphQL resolver duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
	},
	[]string{"operation_type", "resolver_name", "status"},
)

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

func httpMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(wrappedWriter, r)

		route := r.URL.Path
		if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
			routePattern := routeContext.RoutePattern()
			if routePattern != "" {
				route = routePattern
			}
		}

		statusCode := wrappedWriter.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		labels := prometheus.Labels{
			"method":      r.Method,
			"route":       route,
			"status_code": strconv.Itoa(statusCode),
		}

		httpRequestsTotal.With(labels).Inc()
		httpRequestDurationSeconds.With(labels).Observe(time.Since(start).Seconds())
	})
}

func graphqlResolverMetricsMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	start := time.Now()
	operationType := "unknown"
	resolverName := "unknown"

	if operationContext := graphql.GetOperationContext(ctx); operationContext != nil && operationContext.Operation != nil {
		switch operationContext.Operation.Operation {
		case ast.Query:
			operationType = "query"
		case ast.Mutation:
			operationType = "mutation"
		case ast.Subscription:
			operationType = "subscription"
		default:
			operationType = "unknown"
		}
	}

	if fieldContext := graphql.GetFieldContext(ctx); fieldContext != nil {
		switch {
		case fieldContext.Object != "" && fieldContext.Field.Name != "":
			resolverName = fieldContext.Object + "." + fieldContext.Field.Name
		case fieldContext.Field.Name != "":
			resolverName = fieldContext.Field.Name
		}
	}

	defer func() {
		status := "success"
		if recoverValue := recover(); recoverValue != nil {
			status = "error"
			graphqlResolverCallsTotal.WithLabelValues(operationType, resolverName, status).Inc()
			graphqlResolverDurationSeconds.WithLabelValues(operationType, resolverName, status).Observe(time.Since(start).Seconds())
			panic(recoverValue)
		}
		if err != nil {
			status = "error"
		}

		graphqlResolverCallsTotal.WithLabelValues(operationType, resolverName, status).Inc()
		graphqlResolverDurationSeconds.WithLabelValues(operationType, resolverName, status).Observe(time.Since(start).Seconds())
	}()

	res, err = next(ctx)
	return res, err
}
