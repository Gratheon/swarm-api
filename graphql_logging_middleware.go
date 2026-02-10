package main

import (
	"bytes"

	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Gratheon/swarm-api/logger"
)

type gqlRequestBody struct {
	OperationName string `json:"operationName"`
	Query         string `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

func graphqlLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body gqlRequestBody
		if r.Method == "POST" && r.URL.Path == "/graphql" {
			// Read and restore body
			var buf bytes.Buffer
			tee := io.TeeReader(r.Body, &buf)
			data, err := ioutil.ReadAll(tee)
			r.Body.Close()
			r.Body = ioutil.NopCloser(&buf)
			if err == nil {
				_ = json.Unmarshal(data, &body)
			}
			if body.OperationName != "" {
				logger.LogEntrySetField(r, "gql_operation_name", body.OperationName)
			}
			if body.Query != "" {
				opType := "unknown"
				if len(body.Query) > 0 {
					q := body.Query
					if q[0] == 'm' || (len(q) > 7 && q[:8] == "mutation") {
						opType = "mutation"
					} else if q[0] == 'q' || (len(q) > 4 && q[:5] == "query") {
						opType = "query"
					}
				}
				logger.LogEntrySetField(r, "gql_operation_type", opType)
			}
		}
		// User ID from context
		if userID, ok := r.Context().Value("userID").(string); ok {
			logger.LogEntrySetField(r, "user_id", userID)
		}
		next.ServeHTTP(w, r)
	})
}
