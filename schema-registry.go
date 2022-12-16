package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// I'm writing to .version on build time:
//git rev-parse --short HEAD > .version
//go:embed .version
var version string

func RegisterGraphQLSchema(graphqlSchema string, log *logrus.Logger) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Info("Registering schema...")

	service := viper.GetString("schema_registry_url")
	selfUrl := viper.GetString("self_url")

	var tmpVersion = version
	if os.Getenv("ENV_ID") == "dev" {
		tmpVersion = "latest"
	}
	log.Infof("tmpVersion  %s", tmpVersion)

	requestBody, err := json.Marshal(map[string]string{
		"name":      "swarm-api",
		"version":   tmpVersion,
		"url":       selfUrl,
		"type_defs": graphqlSchema,
	})

	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Sending request to to  %v/schema/push", service)

	response, err := client.Post(
		fmt.Sprintf("%v/schema/push", service),
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	log.Infof("schema registry response: %s", response)
	if err != nil {
		log.Error(err)
		return err
	}

	var res map[string]interface{}

	_ = json.NewDecoder(response.Body).Decode(&res)

	log.Infof("schema registry response: %s", response.Body)

	log.Info(res["json"])
	return nil
}
