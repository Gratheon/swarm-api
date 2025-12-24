package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Gratheon/swarm-api/graph"
	"github.com/Gratheon/swarm-api/graph/generated"
	"github.com/Gratheon/swarm-api/logger"
	"github.com/Gratheon/swarm-api/redisPubSub"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

//go:embed schema.graphql
var graphqlSchema string

func main() {
	logger.Info("Starting service")

	logger.Info("Reading config")
	readConfig()

	logger.Info("Initializing logger")
	logrusInstance := logger.InitLogging()

	logger.Info("Initializing redis")
	redisPubSub.InitRedis()

	logger.Info("Initializing router")
	router := chi.NewRouter()

	if os.Getenv("TESTING") != "true" {
		err := RegisterGraphQLSchema(graphqlSchema)
		if err != nil {
			logger.Error("Failed to register schema: " + err.Error())
		}
	} else {
		logger.Info("Skipping schema registration in test mode")
	}

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*", "token"},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	router.Use(authMiddleware)
	//router.Use(logToBugsnag)
	router.Use(logger.NewStructuredLogger(logrusInstance))

	router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))

	serveStaticFiles(router)

	logger.Info("Connecting to DB")
	rootResolver := &graph.Resolver{}
	rootResolver.ConnectToDB()

	gqlGenConfig := generated.Config{Resolvers: rootResolver}
	gqlGenServer := handler.NewDefaultServer(generated.NewExecutableSchema(gqlGenConfig))

	dataLoaderMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := &graph.Loaders{
				HivesByApiaryLoader: graph.NewHiveLoader(rootResolver.Db),
				BoxesByHiveLoader:   graph.NewBoxLoader(rootResolver.Db),
				FamilyByHiveLoader:  graph.NewFamilyLoader(rootResolver.Db),
			}
			ctx := context.WithValue(r.Context(), graph.LoadersKey, loaders)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}

	router.Handle("/graphql", dataLoaderMiddleware(graphqlLoggingMiddleware(gqlGenServer)))

	httpHost := "0.0.0.0:8100"

	log.Printf("Server running on http://%s/graphql", httpHost)

	err := http.ListenAndServe(httpHost, router)

	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func serveStaticFiles(router *chi.Mux) {
	root := "./public"
	fs := http.FileServer(http.Dir(root))

	router.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(root + r.RequestURI)
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
}
