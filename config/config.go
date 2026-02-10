package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecret          string
	JWTExpiryHours     int
	AutoCompleteMinutes int
	ServerPort         string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "taskdb"),
		JWTSecret:          getEnv("JWT_SECRET", "secret-key"),
		JWTExpiryHours:     getEnvInt("JWT_EXPIRY_HOURS", 24),
		AutoCompleteMinutes: getEnvInt("AUTO_COMPLETE_MINUTES", 30),
		ServerPort:         getEnv("SERVER_PORT", "8081"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intVal, _ := strconv.Atoi(value)
	return intVal
}
