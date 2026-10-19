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
	// max_connections, so the pool is sized by config instead of a fixed 25
	viper.SetDefault("db_max_open_conns", 6)
	viper.SetDefault("db_max_idle_conns", 2)
	// Release idle slots quickly (was 2m): other services need them during their bursts
	viper.SetDefault("db_conn_max_idle_time", "30s")

	// database/sql treats MaxOpenConns <= 0 as unlimited, and a bare JSON number becomes nanoseconds:
	// fall back to safe values instead of silently dropping the cap on the shared cluster
	maxOpenConns := viper.GetInt("db_max_open_conns")
	if maxOpenConns < 1 {
		maxOpenConns = 6
	}
	connMaxIdleTime := viper.GetDuration("db_conn_max_idle_time")
	if connMaxIdleTime < time.Second {
		connMaxIdleTime = 30 * time.Second
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(viper.GetInt("db_max_idle_conns"))
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	r.Db = db
}
