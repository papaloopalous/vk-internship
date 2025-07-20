package config

import (
	"log"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile("./config/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
}
