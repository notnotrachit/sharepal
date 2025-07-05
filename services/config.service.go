package services

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/spf13/viper"
)

var Config *models.EnvConfig

func LoadConfig() {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("MODE", "debug")
	v.SetDefault("FIREBASE_CREDENTIALS_JSON", "")
	v.SetConfigType("dotenv")
	v.SetConfigName(".env")
	v.AddConfigPath("./")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	if err := v.Unmarshal(&Config); err != nil {
		panic(err)
	}

	// Debug print for GOOGLE_CLIENT_ID
	googleClientID := v.GetString("GOOGLE_CLIENT_ID")
	println("[DEBUG] GOOGLE_CLIENT_ID from viper:", googleClientID)

	if err := Config.Validate(); err != nil {
		panic(err)
	}
}
