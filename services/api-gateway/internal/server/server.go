package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/config"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/handlers"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/middleware"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func New(cfg *config.Config) *Server {

	//Mode production pour gin
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//Middlewares global
	router.Use(gin.Recovery())
	router.Use(middleware.Cors())
	router.Use(gin.Logger())

	//on va desactivé le ratelimiting en mode debug
	if cfg.RedisURL != "" && !strings.Contains(cfg.RedisURL, "localhost") {
		rateLimiter := middleware.NewRateLimiter(
			cfg.RedisURL,
			cfg.RateLimit,
		)
		router.Use(rateLimiter.RateLimit())
	}

	// Routes
	setupRoutes(router, cfg)

	return &Server{
		router: router,
		config: cfg,
	}
}

func setupRoutes(router *gin.Engine, cfg *config.Config) {

	//Health check
	router.GET("/health", handlers.HealthCheck)

	//API v1
	v1 := router.Group("/api/v1")
	{
		//Authentication
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login(cfg))
			auth.POST("/register", handlers.Register(cfg))
			auth.POST("/validate", handlers.Validate(cfg))
		}

		//services protégés
		protected := v1.Group("/")
		protected.Use(middleware.Auth(cfg.JWT_SECRET))
		{
			//Fichiers
			files := protected.Group("/files")
			{
				files.POST("/upload", handlers.UploadFile(cfg))
				files.GET("/:id", handlers.DownloadFile(cfg))
				files.DELETE("/:id", handlers.DeleteFile(cfg))
			}
		}

	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.config.Port)
}
