package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

func New() *Config {
	cfg := &Config{}

	// Парсим флаги командной строки
	flag.StringVar(&cfg.RunAddress, "a", "", "")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "")
	flag.StringVar(&cfg.SecretKey, "s", "", "")

	// Парсим переменные окружения (перезаписываем значения из флагов, если переменные заданы)
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, proceeding without it")
	}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
