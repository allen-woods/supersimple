package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/allen-woods/supersimple"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	// 	"github.com/allen-woods/supersimple"
)

type Cache struct {
	client redis.UniversalClient
	ttl    time.Duration
}

const apqPrefix = "apq:"
const defaultPort = "8080"
const redisAddr = "localhost:6379"
const redisPass = ""

func NewCache(redisAddress string, password string, ttl time.Duration) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	err := client.Ping().Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Cache{client: client, ttl: ttl}, nil
}

func (c *Cache) Add(ctx context.Context, hash string, query string) {
	c.client.Set(apqPrefix+hash, query, c.ttl)
}

func (c *Cache) Get(ctx context.Context, hash string) (string, bool) {
	s, err := c.client.Get(apqPrefix + hash).Result()
	if err != nil {
		return "", false
	}
	return s, true
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	cache, err := NewCache(redisAddr, redisPass, 24*time.Hour)
	if err != nil {
		log.Fatalf("cannot create APQ redis cache: %v", err)
	}

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(
		supersimple.NewExecutableSchema(supersimple.Config{Resolvers: &supersimple.Resolver{}}),
		handler.EnablePersistedQueryCache(cache),
	))
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// package main

// import (
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/99designs/gqlgen/handler"
// 	"github.com/allen-woods/supersimple"
// )

// const defaultPort = "8080"

// func main() {
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = defaultPort
// 	}

// 	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
// 	http.Handle("/query", handler.GraphQL(supersimple.NewExecutableSchema(supersimple.Config{Resolvers: &supersimple.Resolver{}})))

// 	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
// 	log.Fatal(http.ListenAndServe(":"+port, nil))
//}
