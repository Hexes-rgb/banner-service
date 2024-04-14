package main

import (
	"banner-service/internal/config"
	handlers "banner-service/internal/handlers"
	bannerrepo "banner-service/internal/repositories/banner"
	bannerservice "banner-service/internal/services"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisAddr := redisHost + ":" + redisPort
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	redisCfg := config.RedisConfig{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxConnAge:   30 * time.Minute,
	}

	rdb := config.NewRedisClient(redisCfg)
	defer rdb.Close()

	postgresHost := os.Getenv("DB_HOST")
	postgresPort := os.Getenv("DB_PORT")
	if postgresPort == "" {
		postgresPort = "5432"
	}
	postgresUser := os.Getenv("DB_USER")
	postgresPassword := os.Getenv("DB_PASS")
	postgresDB := os.Getenv("DB_NAME")
	postgresConnStr := "postgresql://" + postgresUser + ":" + postgresPassword + "@" + postgresHost + "/" + postgresDB + "?sslmode=disable"

	postgresCfg := config.PostgresConfig{
		ConnStr:         postgresConnStr,
		MaxConns:        25,
		MaxConnIdleTime: 15 * time.Minute,
	}

	pool, err := config.NewPostgresPool(postgresCfg)
	if err != nil {
		log.Fatalf("Could not create Postgres pool: %v", err)
	}
	defer pool.Close()

	dbRepo := bannerrepo.NewPostgresBannerRepository(pool)

	cacheRepo := bannerrepo.NewRedisBannerRepository(rdb)

	srv := bannerservice.NewBannerService(cacheRepo, dbRepo)

	r := mux.NewRouter()
	handlers.InitBannerRoutes(srv, r)
	handlers.InitUserRoutes(r)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server is starting...")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}
