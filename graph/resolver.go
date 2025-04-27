package graph

import (
	_ "embed"         // Blank import for embed directive
	"encoding/json" // Import encoding/json
	"fmt"
	"math/rand" // Import math/rand
	"time"      // Import time

	"github.com/Gratheon/swarm-api/logger"
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
	// Parse the embedded JSON string during initialization
	err := json.Unmarshal([]byte(femaleNamesJSONString), &r.femaleNamesMap) // Unmarshal into resolver map
	if err != nil {
		logger.Fatal(err.Error()) // Use LogFatal
	}
	// Seed the random number generator once
	rand.Seed(time.Now().UnixNano())

	dsn := viper.GetString("db_dsn")

	logger.Info(fmt.Sprintf("Connecting to DB %s", dsn))
	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		logger.Fatal(err.Error())
	}

	r.Db = db
}
