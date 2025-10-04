package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Load only from .env for local development; ignore if missing
	_ = godotenv.Load("env/.env")
}

type Config struct {
	DatabaseConfig    *DatabaseConfig
	ServerConfig      *ServerConfig
	RedisConfig       *RedisConfig
	EmailConfig       *EmailConfig
	SSLConfig         *SSLConfig
	GoogleCloudConfig *GoogleCloudConfig
	SupabaseConfig    *SupabaseConfig
}

func NewConfig() *Config {
	return &Config{
		DatabaseConfig:    NewDatabaseConfig(),
		ServerConfig:      NewServerConfig(),
		RedisConfig:       NewRedisConfig(),
		EmailConfig:       NewEmailConfig(),
		SSLConfig:         NewSSLConfig(),
		GoogleCloudConfig: NewGoogleCloudConfig(),
		SupabaseConfig:    NewSupabaseConfig(),
	}
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
}

type ServerConfig struct {
	Environment string
	Host        string
	Port        string
	SecretKey   string
	ApiSecret   string
}

func NewServerConfig() *ServerConfig {
	port := os.Getenv("PORT_NUMBER")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	return &ServerConfig{
		Environment: os.Getenv("ENVIRONMENT"),
		Host:        host,
		Port:        port,
		SecretKey:   os.Getenv("SECRET_KEY"),
		ApiSecret:   os.Getenv("API_SECRET"),
	}
}

type SSLConfig struct {
	SSLEnabled bool
	CertFile   string
	KeyFile    string
}

func NewSSLConfig() *SSLConfig {
	return &SSLConfig{
		SSLEnabled: os.Getenv("SSL_ENABLED") == "true",
		CertFile:   os.Getenv("SSL_CERT_FILE"),
		KeyFile:    os.Getenv("SSL_KEY_FILE"),
	}
}

type RedisConfig struct {
	Address  string
	DB       int
	Username string
	Password string
}

func NewRedisConfig() *RedisConfig {
	if os.Getenv("REDIS_DB") == "" {
		os.Setenv("REDIS_DB", "0")
	}
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Error converting REDIS_DB to int: %v", err)
	}
	return &RedisConfig{
		Address:  os.Getenv("REDIS_ADDRESS"),
		DB:       db,
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     os.Getenv("FROM_NAME"),
	}
}

type GoogleCloudConfig struct {
	GoogleClientId     string
	GoogleClientSecret string
	GoogleRedirectUris string
}

func NewGoogleCloudConfig() *GoogleCloudConfig {
	return &GoogleCloudConfig{
		GoogleClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectUris: os.Getenv("GOOGLE_REDIRECT_URIS"),
	}
}

type SupabaseConfig struct {
	URL        string
	Key        string
	Bucket     string
	ServiceKey string
}

func NewSupabaseConfig() *SupabaseConfig {
	return &SupabaseConfig{
		URL:        os.Getenv("SUPABASE_URL"),
		Key:        os.Getenv("SUPABASE_KEY_ROLE"),
		Bucket:     os.Getenv("SUPABASE_BUCKET"),
		ServiceKey: os.Getenv("SUPABASE_SERVICE_KEY"),
	}
}
