package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	logger "github.com/Gratheon/log-lib-go"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	dbUnavailableMessage = "Database is temporarily unavailable, please retry"
	internalErrorMessage = "Internal server error"
	dbUnavailableCode    = "DATABASE_UNAVAILABLE"
	internalErrorCode    = "INTERNAL_SERVER_ERROR"
)

// Fallback for errors that were flattened with %v/%s before reaching the presenter:
// driver and network messages carry the DB host, IP, user and database name.
var leakyErrorMarkers = []string{
	"failed to connect to",
	"SQLSTATE",
	"dial tcp",
	"connection refused",
	"pgconn",
	"pgx",
	"ondigitalocean.com",
}

// graphqlErrorPresenter keeps resolver errors as they are, except for database and
// network errors: those are logged in full server-side and replaced with a generic
// message, so clients never see DB hosts, IPs, usernames or SQL.
func graphqlErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)

	kind := classifyInternalError(err)
	if kind == "" {
		return gqlErr
	}

	logger.ErrorWithContext(ctx, "GraphQL resolver failed with internal error at "+gqlErr.Path.String()+": "+err.Error())

	message, code := internalErrorMessage, internalErrorCode
	if kind == dbUnavailableCode {
		message, code = dbUnavailableMessage, dbUnavailableCode
	}
	return &gqlerror.Error{
		Message:    message,
		Path:       gqlErr.Path,
		Locations:  gqlErr.Locations,
		Extensions: map[string]interface{}{"code": code},
	}
}

// classifyInternalError returns "" for errors that are safe to show to clients,
// dbUnavailableCode for connection problems and internalErrorCode for other DB errors.
func classifyInternalError(err error) string {
	if err == nil {
		return ""
	}

	var connectErr *pgconn.ConnectError
	var netErr net.Error
	if errors.As(err, &connectErr) || errors.As(err, &netErr) ||
		errors.Is(err, sql.ErrConnDone) || errors.Is(err, driver.ErrBadConn) {
		return dbUnavailableCode
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if isConnectionPgErrorCode(pgErr.Code) {
			return dbUnavailableCode
		}
		return internalErrorCode
	}

	message := err.Error()
	for _, marker := range leakyErrorMarkers {
		if strings.Contains(message, marker) {
			if strings.Contains(message, "failed to connect to") || strings.Contains(message, "53300") {
				return dbUnavailableCode
			}
			return internalErrorCode
		}
	}
	return ""
}

// isConnectionPgErrorCode matches SQLSTATE classes that mean "retry later":
// 08 connection exception, 53 insufficient resources (incl. 53300 too_many_connections),
// 57P admin shutdown / cannot connect now.
func isConnectionPgErrorCode(code string) bool {
	return strings.HasPrefix(code, "08") || strings.HasPrefix(code, "53") || strings.HasPrefix(code, "57P")
}
