package main

import (
	"fmt"
	"log"

	"travia.backend/api/handler"
	"travia.backend/config"
	db "travia.backend/db/sqlc"
)

// @title           Travia API
// @version         1.0
// @description     Travia API for travel management
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @responseHeader 200 {string} X-Response-Time "Response time in milliseconds"

func init() {
	config.LoadEnv()

}

func main() {
	// Load configuration
	config := config.NewConfig()

	// Initialize database connection
	conn, err := db.InitDB(config.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB(conn)

	// Initialize Travia database interface
	travia := db.NewTravia(conn)

	// Initialize Redis client
	redisClient := db.InitRedis(config.RedisConfig)
	defer redisClient.Close()

	log.Printf("Server starting on %s:%s", config.ServerConfig.Host, config.ServerConfig.Port)

	server := handler.NewServer(config, travia, redisClient)
	server.Start(fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port))
}
