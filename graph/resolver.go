package graph

import (
	_ "embed"       // Blank import for embed directive
	"encoding/json" // Import encoding/json
	"math/rand"     // Import math/rand
	"time"          // Import time

	"github.com/Gratheon/log-lib-go"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

//go:embed female-names.json
var femaleNamesJSONString string // Embed as string from the same directory

//go:generate go run github.com/99designs/gqlgen -v

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Db             *sqlx.DB
	femaleNamesMap map[string][]string // Add map to resolver
}

func (r *Resolver) ConnectToDB() {
	err := json.Unmarshal([]byte(femaleNamesJSONString), &r.femaleNamesMap)
	if err != nil {
		logger.Fatal(err.Error())
	}
	rand.Seed(time.Now().UnixNano())

	dsn := viper.GetString("db_dsn")

	logger.Info("Connecting to DB")
	db, err := OpenDB(dsn)

	if err != nil {
		logger.Fatal(err.Error())
	}

	// The managed Postgres cluster is shared by several services and has a small
	// max_connections, so this single pool is the whole connection budget of the service.
	// When all connections are busy, parallel GraphQL requests wait for a free one
	// instead of opening more. Idle connections are kept up to the cap so bursts reuse
	// warm connections instead of reconnecting (each reconnect can hit a full server).
	viper.SetDefault("db_max_open_conns", 14)
	maxOpenConns := viper.GetInt("db_max_open_conns")
	viper.SetDefault("db_max_idle_conns", maxOpenConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(viper.GetInt("db_max_idle_conns"))
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	r.Db = db
}
