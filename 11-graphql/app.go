package main

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"stepikGoWebServices/graph"
	"stepikGoWebServices/graph/model"
	"stepikGoWebServices/graph/resolvers"
)

func GetApp() http.Handler {
	data, err := os.ReadFile("testdata.json")
	if err != nil {
		log.Fatal("Error reading testdata.json:", err)
	}

	var testData resolvers.TestData
	err = json.Unmarshal(data, &testData)
	if err != nil {
		log.Fatal("Error parsing testdata.json:", err)
	}

	var resolver = &resolvers.Resolver{
		Data:       testData,
		UserCarts:  make(map[string][]model.CartItem),
		UserTokens: make(map[string]string),
	}

	graphQlServer := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: resolver,
			},
		),
	)

	router := mux.NewRouter()

	router.Handle("/query", resolver.AuthMiddleware(graphQlServer))
	router.HandleFunc("/register", resolver.RegisterHandler).Methods("POST")

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))

	return router
}
