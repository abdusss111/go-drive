package server

import (
	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/logger"
	"github.com/abduss/godrive/internal/metrics"
	"github.com/abduss/godrive/internal/middleware"
	"github.com/abduss/godrive/internal/presigned"
	"github.com/abduss/godrive/internal/usage"
	_ "github.com/abduss/godrive/docs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Dependencies groups the services required by the HTTP router.
type Dependencies struct {
	Config        config.Config
	DB            *pgxpool.Pool
	ObjectStore   *minio.Client
	AuthService      *auth.Service
	BucketService    *bucket.Service
	FileService      *file.Service
	UsageService     *usage.Service
	PresignedService *presigned.Service
	Logger           *logger.Logger
}

// NewRouter builds a Gin engine with foundational middleware and routes.
func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Enable CORS for cross-origin requests
	router.Use(middleware.CORS())
	
	// Use structured logging with correlation IDs if logger is provided
	if deps.Logger != nil {
		router.Use(logger.CorrelationIDMiddleware(deps.Logger))
	} else {
		router.Use(gin.Logger())
	}

	// Register metrics middleware if available
	if deps.Config.Metrics.PrometheusPath != "" {
		metrics.InitMetrics()
		router.Use(metrics.Middleware())
		metrics.Register(router, deps.Config.Metrics.PrometheusPath)
	}

	registerHealthRoutes(router, deps)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/v1")
	if deps.AuthService != nil {
		auth.RegisterRoutes(api, deps.AuthService)

		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware(deps.AuthService))

		if deps.BucketService != nil {
			bucket.RegisterRoutes(protected, deps.BucketService)
		}
		if deps.FileService != nil {
			file.RegisterRoutes(protected, deps.FileService)
		}
		if deps.UsageService != nil && deps.BucketService != nil {
			usage.RegisterRoutes(protected, deps.UsageService, deps.BucketService)
		}
		if deps.PresignedService != nil {
			presigned.RegisterRoutes(protected, deps.PresignedService)
		}
	}

	return router
}
