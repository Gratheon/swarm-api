package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Gratheon/log-lib-go"
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

	response, err := client.Post(
		fmt.Sprintf("%v/schema/push", service),
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	defer response.Body.Close()

	var res map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse schema registry response: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Schema registered successfully (version: %s, status: %s)", tmpVersion, response.Status))
	return nil
}
