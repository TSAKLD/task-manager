package bootstrap

import (
	"errors"
	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASS"`
	DBName     string `env:"DB_NAME"`
	HTTPPort   string `env:"HTTP_PORT"`
	RedisAddr  string `env:"REDIS_ADDR"`
	KafkaAddr  string `env:"KAFKA_ADDR"`
	KafkaTopic string `env:"KAFKA_TOPIC"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	var config Config

	_, err = env.UnmarshalFromEnviron(&config)
	if err != nil {
		log.Fatal(err)
	}

	return &config, nil
}

func (c *Config) Validate() []error {
	var errorList []error

	if c.DBHost == "" {
		err := errors.New("invalid DB host field \n")
		errorList = append(errorList, err)
	}

	if c.DBPort == "" {
		err := errors.New("invalid DB port field \n")
		errorList = append(errorList, err)
	}

	if c.DBUser == "" {
		err := errors.New("invalid DB user field \n")
		errorList = append(errorList, err)
	}

	if c.DBPassword == "" {
		err := errors.New("invalid DB password field \n")
		errorList = append(errorList, err)
	}

	if c.DBName == "" {
		err := errors.New("invalid DB name field \n")
		errorList = append(errorList, err)
	}

	if c.HTTPPort == "" {
		err := errors.New("invalid HTTP port field \n")
		errorList = append(errorList, err)
	}

	if c.RedisAddr == "" {
		err := errors.New("invalid Redis host field \n")
		errorList = append(errorList, err)
	}

	if c.KafkaAddr == "" {
		err := errors.New("invalid Kafka address field \n")
		errorList = append(errorList, err)
	}

	if c.KafkaTopic == "" {
		err := errors.New("invalid Kafka topic field \n")
		errorList = append(errorList, err)
	}

	if len(errorList) != 0 {
		return errorList
	}

	return nil
}
