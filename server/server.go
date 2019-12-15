package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/allen-woods/supersimple"
	"github.com/allen-woods/supersimple/auth"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/rs/cors"

	"github.com/pkg/errors"
)

type Cache struct {
	client redis.UniversalClient
	ttl    time.Duration
}

const apqPrefix = "apq:"
const defaultPort = "8080"
const redisAddr = "localhost:6379"
const redisPass = ""

var roller struct {
	t *time.Ticker
	d chan bool
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}

func NewRedisCache(client *redis.Client, redisAddress string, password string, ttl time.Duration) (*Cache, error) {
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

func init() {
	// Initial trigger, or we'll be waiting an hour for this to fire.
	err := auth.Roll()
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	// This goroutine will fire once every hour, continuously.
	roller.t = time.NewTicker(4 * time.Second)
	roller.d = make(chan bool)

	go func() {
		for {
			select {
			case <-roller.d:
				return
			case <-roller.t.C:
				err := auth.Roll()
				if err != nil {
					fmt.Println("ERROR:", err)
				}
			}
		}
	}()

	fmt.Println("Roller is running.")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	redisClient := NewRedisClient()

	cache, err := NewRedisCache(redisClient, redisAddr, redisPass, 24*time.Hour)
	if err != nil {
		log.Fatalf("cannot create APQ redis cache: %v", err)
	}

	router := chi.NewRouter()

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)
	router.Use(auth.Middleware())

	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	router.Handle("/query", handler.GraphQL(
		supersimple.NewExecutableSchema(supersimple.Config{Resolvers: &supersimple.Resolver{}}),
		handler.EnablePersistedQueryCache(cache),
	))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	err = http.ListenAndServe(":"+port,
		// csrf.Protect(
		// 	[]byte("32-byte-long-auth-key"),
		// 	csrf.Secure(false),
		// 	csrf.CookieName("_csrf"),
		// )(
		router,
	)
	if err != nil {
		roller.t.Stop()
		roller.d <- true
		panic(err)
	}
}
