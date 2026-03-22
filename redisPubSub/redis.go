package redisPubSub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Gratheon/log-lib-go"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type Payload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var ctx = context.Background()
var client *redis.Client

func InitRedis() *redis.Client {
	redisAddress := viper.GetString("redis_address")
	logger.Info("Connecting to Redis at " + redisAddress)
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: viper.GetString("redis_pass"),
		DB:       0,
	})

	return client
}

func PublishEvent(uid string, entity string, entityID string, verb string, data interface{}) {
	if client == nil {
		client = InitRedis()
	}
	payloadJSON, _ := json.Marshal(data)
	channel := fmt.Sprintf("%s.%s.%s.%s", uid, entity, entityID, verb)

	err := client.Publish(ctx, channel, payloadJSON).Err()

	if err != nil {
		logger.Error(fmt.Sprintf("Redis publish error on channel %s: %v", channel, err))
	}

	err = client.Publish(ctx, channel+"."+verb, payloadJSON).Err()

	if err != nil {
		logger.Error(fmt.Sprintf("Redis publish error on channel %s.%s: %v", channel, verb, err))
	}
}
