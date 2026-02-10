package main

import (
	"fmt"
	"os"
	"github.com/spf13/viper"
)

func readConfig() {
	configFileName :="config." + os.Getenv("ENV_ID")

	if os.Getenv("NATIVE") == "1"{
		configFileName = "config.native"
	}

	viper.AddConfigPath("./config")
	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.SetConfigType("json")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // path to look for the config file in

	viper.SafeWriteConfig()

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
