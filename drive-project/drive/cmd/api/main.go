package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/logger"
	"github.com/abduss/godrive/internal/presigned"
	"github.com/abduss/godrive/internal/server"
	"github.com/abduss/godrive/internal/storage"
	"github.com/abduss/godrive/internal/usage"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Initialize structured logger
	appLogger, err := logger.Init()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer appLogger.Sync()

	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("load config", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		appLogger.Fatal("connect postgres", err)
	}
	defer dbPool.Close()

	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		appLogger.Fatal("connect minio", err)
	}

	if err := storage.EnsureBucket(ctx, minioClient, cfg.MinIO.Bucket, cfg.MinIO.Region); err != nil {
		appLogger.Fatal("ensure bucket", err)
		return
	}

	authRepo := auth.NewRepository(dbPool)
	authService := auth.NewService(authRepo, cfg.Auth)

	bucketRepo := bucket.NewRepository(dbPool)
	fileRepo := file.NewRepository(dbPool)

	usageRepo := usage.NewRepository(dbPool)
	usageService := usage.NewService(usageRepo)

	bucketService := bucket.NewService(bucketRepo, fileRepo, minioClient, cfg.MinIO.Bucket)
	fileStore := file.NewMinIOStore(minioClient)
	fileService := file.NewService(fileRepo, bucketRepo, fileStore, cfg.MinIO.Bucket, usageService)
	
	// Create presigned service with adapters
	bucketAdapter := &presigned.BucketServiceAdapter{Service: bucketService}
	fileAdapter := &presigned.FileRepositoryAdapter{Repo: fileRepo}
	minioAdapter := &presigned.MinIOClientAdapter{Client: minioClient}
	presignedService := presigned.NewService(bucketAdapter, fileAdapter, minioAdapter, cfg.MinIO.Bucket)

	router := server.NewRouter(server.Dependencies{
		Config:           cfg,
		DB:               dbPool,
		ObjectStore:      minioClient,
		AuthService:      authService,
		BucketService:    bucketService,
		FileService:      fileService,
		UsageService:     usageService,
		PresignedService: presignedService,
		Logger:           appLogger,
	})

	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		appLogger.Info("GoDrive API listening", zap.String("address", cfg.Server.Address()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("http server", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	appLogger.Info("shutting down gracefully")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("shutdown error", err)
	}
}
