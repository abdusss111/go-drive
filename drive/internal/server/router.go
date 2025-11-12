package server

import (
	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

// Dependencies groups the services required by the HTTP router.
type Dependencies struct {
	Config        config.Config
	DB            *pgxpool.Pool
	ObjectStore   *minio.Client
	AuthService   *auth.Service
	BucketService *bucket.Service
	FileService   *file.Service
}

// NewRouter builds a Gin engine with foundational middleware and routes.
func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

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
	}

	return router
}
