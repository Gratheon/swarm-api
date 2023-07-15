package graph

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/Gratheon/swarm-api/logger"
)

//go:generate go run github.com/99designs/gqlgen -v

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Db *sqlx.DB
}

func (r *Resolver) ConnectToDB() {
	dsn := viper.GetString("db_dsn")

	logger.LogInfo(fmt.Sprintf("Connecting to DB %s", dsn))
	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		logger.LogFatal(err)
	}

	r.Db = db
}
