package server

import (
	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/metrics"
	"github.com/abduss/godrive/internal/presigned"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// Dependencies groups the services required by the HTTP router.
type Dependencies struct {
	Config           config.Config
	Logger           *zap.Logger
	DB               *pgxpool.Pool
	ObjectStore      *minio.Client
	AuthService      *auth.Service
	BucketService    *bucket.Service
	FileService      *file.Service
	PresignedService *presigned.Service
}

// NewRouter builds a Gin engine with foundational middleware and routes.
func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(loggerMiddleware(deps.Logger))
	router.Use(securityHeadersMiddleware(deps.Config.Security))
	if deps.Config.RequestLimits.MaxBodyBytes > 0 {
		router.Use(limitBodyMiddleware(deps.Config.RequestLimits.MaxBodyBytes))
	}
	if deps.Config.RateLimit.Enabled {
		router.Use(rateLimitMiddleware(deps.Config.RateLimit))
	}
	if deps.Config.CORS.Enabled {
		router.Use(newCORS(deps.Config.CORS))
	}

	registerHealthRoutes(router, deps)
	metrics.Register(router, deps.Config.Metrics.PrometheusPath)

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
		if deps.PresignedService != nil {
			presigned.RegisterRoutes(protected, deps.PresignedService)
		}
	}

	return router
}
