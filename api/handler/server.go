package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"travia.backend/config"
	db "travia.backend/db/sqlc"
)

type Server struct {
	config *config.Config
	router *gin.Engine
	z      db.Z
	redis  *redis.Client
}

func NewServer(config *config.Config, z db.Z, redisClient *redis.Client) *Server {
	server := &Server{
		config: config,
		z:      z,
		redis:  redisClient,
		router: gin.Default(),
	}

	// Disable trailing slash redirects to avoid 301 redirects
	server.router.RedirectTrailingSlash = false
	server.router.RedirectFixedPath = false

	// Setup router components
	server.SetupMiddlewares()
	server.SetupAuthProviders()
	//server.InitStripe() // Initialize Stripe
	server.SetupRoutes()
	server.SetupSwagger()

	return server
}

func (s *Server) Start(address string) error {
	if s.config.SSLConfig.SSLEnabled {
		log.Printf("Starting HTTPS server on %s", address)
		return s.router.RunTLS(address, s.config.SSLConfig.CertFile, s.config.SSLConfig.KeyFile)
	}
	log.Printf("Starting HTTP server on %s", address)
	if s.config.ServerConfig.Port == "" {
		return s.router.Run(":8080")
	}
	return s.router.Run(address)
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
