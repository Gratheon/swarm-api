package model

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// Postgres has no LastInsertId(): inserts that need the new id end with "RETURNING id"
// and read it back with these helpers. Both *sqlx.DB and *sqlx.Tx can be passed.

func insertReturningID(db sqlx.Queryer, query string, args ...interface{}) (int64, error) {
	var id int64
	err := db.QueryRowx(query, args...).Scan(&id)
	return id, err
}

// namedInsertReturningID is insertReturningID for queries with :name parameters (former NamedExec).
func namedInsertReturningID(db sqlx.Queryer, query string, arg interface{}) (int64, error) {
	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, arg)
	if err != nil {
		return 0, err
	}
	return insertReturningID(db, boundQuery, args...)
}

const pgUniqueViolation = "23505"

// isUniqueViolation reports a duplicate key error (MySQL 1062 equivalent).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
