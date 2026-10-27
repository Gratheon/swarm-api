//go:build !integration
// +build !integration

package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

const leakedConnectError = "failed to connect to `user=swarm_api database=swarm_api`: 167.172.42.83:25060 " +
	"(db-postgresql-ams3-82076-do-user-305335-0.i.db.ondigitalocean.com): server error: " +
	"FATAL: remaining connection slots are reserved for roles with the SUPERUSER attribute (SQLSTATE 53300)"

func presentInField(err error) (message string, code interface{}, path string) {
	ctx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
		Field: graphql.CollectedField{Field: &ast.Field{Alias: "hive"}},
	})
	gqlErr := graphqlErrorPresenter(ctx, err)
	return gqlErr.Message, gqlErr.Extensions["code"], gqlErr.Path.String()
}

func TestGraphqlErrorPresenter_HidesConnectionDetails(t *testing.T) {
	cases := map[string]error{
		"pg too many connections":  fmt.Errorf("get hive: %w", &pgconn.PgError{Code: "53300", Message: "remaining connection slots are reserved"}),
		"flattened connect error":  errors.New(leakedConnectError),
		"connect error wrapped %v": fmt.Errorf("db: %v", errors.New(leakedConnectError)),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			message, code, path := presentInField(err)
			assert.Equal(t, dbUnavailableMessage, message)
			assert.Equal(t, dbUnavailableCode, code)
			assert.Equal(t, "hive", path)
			assert.NotContains(t, message, "167.172.42.83")
			assert.NotContains(t, message, "swarm_api")
		})
	}
}

func TestGraphqlErrorPresenter_HidesQueryErrors(t *testing.T) {
	message, code, _ := presentInField(&pgconn.PgError{Code: "42P01", Message: `relation "hives" does not exist`})
	assert.Equal(t, internalErrorMessage, message)
	assert.Equal(t, internalErrorCode, code)
}

func TestGraphqlErrorPresenter_KeepsApplicationErrors(t *testing.T) {
	message, code, _ := presentInField(errors.New("hive limit reached for your billing plan"))
	assert.Equal(t, "hive limit reached for your billing plan", message)
	assert.Nil(t, code)
}
