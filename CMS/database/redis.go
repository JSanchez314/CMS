package database

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisonce   sync.Once
)

func RedisInit() {
	redisonce.Do(func() {
		addr := os.Getenv("ADDRe")
		password := os.Getenv("PW")
		dbString := os.Getenv("DB")

		db, err := strconv.Atoi(dbString)
		if err != nil {
			log.Fatal("Error converting db to int: ", err)
		}

		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = redisClient.Ping(ctx).Result()
		if err != nil {
			log.Fatal("Error connecting to redis: ", err)
		}

		log.Println("Connected to redis")
	})
}

func RedisGetClient() *redis.Client {
	if redisClient == nil {
		log.Fatal("Redis is not initialized. Call RedisInit() first.")
	}
	return redisClient
}
