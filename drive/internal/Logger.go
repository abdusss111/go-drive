package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/abduss/godrive/internal/auth"
	"github.github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/logger"
	"github.com/abduss/godrive/internal/server"
	"github.com/abduss/godrive/internal/storage"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	logg, err := logger.Init()
	if err != nil {
		panic("init logger: " + err.Error())
	}
	defer logg.Sync()

	cfg, err := config.Load()
	if err != nil {
		logg.Fatal("load config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		logg.Fatal("connect postgres", zap.Error(err))
	}
	defer dbPool.Close()

	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		logg.Fatal("connect minio", zap.Error(err))
	}

	if err := storage.EnsureBucket(ctx, minioClient, cfg.MinIO.Bucket); err != nil {
		logg.Fatal("ensure bucket", zap.Error(err))
	}

	authRepo := auth.NewRepository(dbPool)
	authService := auth.NewService(authRepo, cfg.Auth)

	bucketRepo := bucket.NewRepository(dbPool)
	fileRepo := file.NewRepository(dbPool)

	bucketService := bucket.NewService(bucketRepo, fileRepo, minioClient, cfg.MinIO.Bucket)
	fileStore := file.NewMinIOStore(minioClient)
	fileService := file.NewService(fileRepo, bucketRepo, fileStore, cfg.MinIO.Bucket)

	router := server.NewRouter(server.Dependencies{
		Config:        cfg,
		DB:            dbPool,
		ObjectStore:   minioClient,
		AuthService:   authService,
		BucketService: bucketService,
		FileService:   fileService,
	})

	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logg.Info("GoDrive API listening", zap.String("address", cfg.Server.Address()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Fatal("http server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logg.Info("shutting down gracefully")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logg.Error("shutdown error", zap.Error(err))
	}
}
