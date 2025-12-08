package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Gratheon/swarm-api/logger"
	"github.com/spf13/viper"
)

// I'm writing to .version on build time:
// git rev-parse --short HEAD > .version
//
//go:embed .version
var version string

func RegisterGraphQLSchema(graphqlSchema string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	logger.Info("Registering schema...")

	service := viper.GetString("schema_registry_url")
	selfUrl := viper.GetString("self_url")

	var tmpVersion = version
	if os.Getenv("ENV_ID") == "dev" {
		tmpVersion = "latest"
	}
	logger.Info(fmt.Sprintf("tmpVersion  %s", tmpVersion))

	requestBody, err := json.Marshal(map[string]string{
		"name":      "swarm-api",
		"version":   tmpVersion,
		"url":       selfUrl,
		"type_defs": graphqlSchema,
	})

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info(fmt.Sprintf("Sending request to to  %v/schema/push", service))

	response, err := client.Post(
		fmt.Sprintf("%v/schema/push", service),
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info(fmt.Sprintf("schema registry response status: %s", response.Status))

	var res map[string]interface{}

	_ = json.NewDecoder(response.Body).Decode(&res)

	logger.Info(fmt.Sprintf("schema registry response: %s", response.Body))

	if jsonMsg, ok := res["json"].(string); ok {
		logger.Info(jsonMsg)
	} else {
		logger.Info(fmt.Sprintf("schema registry response json: %v", res["json"]))
	}
	return nil
}
