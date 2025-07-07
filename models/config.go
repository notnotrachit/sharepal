package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type EnvConfig struct {
	ServerPort                 string `mapstructure:"SERVER_PORT"`
	ServerAddr                 string `mapstructure:"SERVER_ADDR"`
	MongodbUri                 string `mapstructure:"MONGO_URI"`
	MongodbDatabase            string `mapstructure:"MONGO_DATABASE"`
	UseRedis                   bool   `mapstructure:"USE_REDIS"`
	RedisDefaultAddr           string `mapstructure:"REDIS_DEFAULT_ADDR"`
	JWTSecretKey               string `mapstructure:"JWT_SECRET"`
	JWTAccessExpirationMinutes int    `mapstructure:"JWT_ACCESS_EXPIRATION_MINUTES"`
	JWTRefreshExpirationDays   int    `mapstructure:"JWT_REFRESH_EXPIRATION_DAYS"`
	Mode                       string `mapstructure:"MODE"`
	VapidPublicKey             string `mapstructure:"VAPID_PUBLIC_KEY"`
	VapidPrivateKey            string `mapstructure:"VAPID_PRIVATE_KEY"`
	GoogleClientID             string `mapstructure:"GOOGLE_CLIENT_ID"`
	AWSRegion                  string `mapstructure:"AWS_REGION"`
	AWSS3Bucket                string `mapstructure:"AWS_S3_BUCKET"`
	AWSAccessKeyID             string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey         string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSS3Endpoint              string `mapstructure:"AWS_S3_ENDPOINT"`
}

func (config *EnvConfig) Validate() error {
	return validation.ValidateStruct(config,
		validation.Field(&config.ServerPort, is.Port),
		validation.Field(&config.ServerAddr, validation.Required),

		validation.Field(&config.MongodbUri, validation.Required),
		validation.Field(&config.MongodbDatabase, validation.Required),

		validation.Field(&config.UseRedis, validation.In(true, false)),
		validation.Field(&config.RedisDefaultAddr),

		validation.Field(&config.JWTSecretKey, validation.Required),
		validation.Field(&config.JWTAccessExpirationMinutes, validation.Required),
		validation.Field(&config.JWTRefreshExpirationDays, validation.Required),

		validation.Field(&config.Mode, validation.In("debug", "release")),
		validation.Field(&config.VapidPublicKey, validation.Required),
		validation.Field(&config.VapidPrivateKey, validation.Required),
		validation.Field(&config.GoogleClientID, validation.Required),
	)
}
