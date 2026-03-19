package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gratheon/swarm-api/logger"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQLLoggingMiddleware_SetsOperationFields(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		expectedType  string
		operationName string
	}{
		{name: "mutation", query: "mutation { createHive { id } }", expectedType: "mutation", operationName: "CreateHive"},
		{name: "query", query: "query { hives { id } }", expectedType: "query", operationName: "ListHives"},
		{name: "unknown", query: "{ hives { id } }", expectedType: "unknown", operationName: "Anonymous"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"operationName":"` + tc.operationName + `","query":"` + tc.query + `","variables":{}}`
			req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			entry := &logger.StructuredLoggerEntry{Logger: logrus.NewEntry(logrus.New())}
			ctx := context.WithValue(req.Context(), chimiddleware.LogEntryCtxKey, entry)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := graphqlLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusNoContent, rr.Code)
			logEntry, ok := entry.Logger.(*logrus.Entry)
			require.True(t, ok, "expected logger to be a logrus.Entry")
			assert.Equal(t, tc.operationName, logEntry.Data["gql_operation_name"])
			assert.Equal(t, tc.expectedType, logEntry.Data["gql_operation_type"])
		})
	}
}

func TestGraphQLLoggingMiddleware_SetsUserIDFromContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	entry := &logger.StructuredLoggerEntry{Logger: logrus.NewEntry(logrus.New())}
	ctx := context.WithValue(req.Context(), chimiddleware.LogEntryCtxKey, entry)
	ctx = context.WithValue(ctx, "userID", "user-1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler := graphqlLoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	logEntry, ok := entry.Logger.(*logrus.Entry)
	require.True(t, ok, "expected logger to be a logrus.Entry")
	assert.Equal(t, "user-1", logEntry.Data["user_id"])
	assert.Nil(t, logEntry.Data["gql_operation_name"])
	assert.Nil(t, logEntry.Data["gql_operation_type"])
}
