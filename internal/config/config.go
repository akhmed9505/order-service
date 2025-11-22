package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DB struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     int
}

type Kafka struct {
	Brokers string
	Topic   string
	GroupID string
}

type Config struct {
	DBcfg    DB
	Kafkacfg Kafka
	HTTPPort string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	portStr := os.Getenv("DB_PORT")
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	db := DB{
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     portInt,
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS env is required")
	}

	kafkaCfg := Kafka{
		Brokers: kafkaBrokers,
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	return &Config{
		DBcfg:    db,
		Kafkacfg: kafkaCfg,
		HTTPPort: os.Getenv("HTTP_PORT"),
	}
}
