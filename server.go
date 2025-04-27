package main

import (
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
	log.Print("Starting service")

	log.Print("Reading config")
	readConfig()

	log.Print("Initializing logger")
	logrusInstance := logger.InitLogging()

	log.Print("Initializing redis")
	redisPubSub.InitRedis()

	log.Print("Initializing router")
	router := chi.NewRouter()

	_ = RegisterGraphQLSchema(graphqlSchema, logrusInstance)

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

	log.Print("Connecting to DB")
	rootResolver := &graph.Resolver{}
	rootResolver.ConnectToDB()

	gqlGenConfig := generated.Config{Resolvers: rootResolver}
	gqlGenServer := handler.NewDefaultServer(generated.NewExecutableSchema(gqlGenConfig))
	router.Handle("/graphql", gqlGenServer)

	httpHost := "0.0.0.0:8100"

	err := http.ListenAndServe(httpHost, router)

	log.Printf("Server running on http://%s:%s/graphql", httpHost)

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
