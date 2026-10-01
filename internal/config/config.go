package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	DBDriver    string
	SSHPort     string
	JWTSecret   string
	StripeKey   string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "./ecommerce.db"
	}

	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-change-in-production"
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")

	return &Config{
		ServerPort:  port,
		DatabaseURL: dbURL,
		DBDriver:    dbDriver,
		SSHPort:     sshPort,
		JWTSecret:   jwtSecret,
		StripeKey:   stripeKey,
	}
}
