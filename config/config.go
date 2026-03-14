package config

import "os"

type Config struct {
	DatabaseURL    string
	DatabaseSchema string // Add schema if needed
	MongoURI       string
	MongoDBName    string
	ServerAddress  string
	JWTSecretKey   string
	HashPepper     string
	SuperVendorID  string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DatabaseSchema: getEnv("DATABASE_SCHEMA", "public"), // If using a specific schema
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGO_DB_NAME", "fube_local"),
		ServerAddress:  getEnv("SERVER_ADDRESS", port),
		JWTSecretKey:   getEnv("JWT_SECRET_KEY", "default-jwt-secret-keep-it-safe"),
		HashPepper:     getEnv("HASH_PEPPER", ""),
		SuperVendorID:  getEnv("SUPER_VENDOR_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
