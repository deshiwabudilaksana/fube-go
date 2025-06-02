package config

import "os"

type Config struct {
	DatabaseURL    string
	DatabaseSchema string // Add schema if needed
	ServerAddress  string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://info.gondry:FaOum4PA9RMY@ep-flat-sunset-89661848.ap-southeast-1.aws.neon.tech/fube?sslmode=require"),
		DatabaseSchema: getEnv("DATABASE_SCHEMA", "public"), // If using a specific schema
		ServerAddress:  getEnv("SERVER_ADDRESS", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
