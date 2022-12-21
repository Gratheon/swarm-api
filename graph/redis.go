package graph

import (
	"encoding/json"
	"context"
	"fmt"
	"github.com/spf13/viper"
    "github.com/go-redis/redis/v8"
)

type Payload struct {
	ID    string `json:"id"`
	Name string `json:"name"`
}

var ctx = context.Background()
var client *redis.Client

func InitRedis() *redis.Client {
	redisAddress := viper.GetString("redis_address")
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: viper.GetString("redis_pass"),
		DB:       0,
	})

	return client;
}

func PublishEvent(channel string, data interface{}){
	if client == nil {
		client = InitRedis();
	}
	payloadJSON, _ := json.Marshal(data)
	
	fmt.Printf("redis publish %v", payloadJSON)

	err := client.Publish(ctx, channel, payloadJSON).Err()

	if err != nil {
		fmt.Printf("redis publish error %v", err)
	}
}