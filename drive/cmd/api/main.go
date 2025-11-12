package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/config"
	"github.com/abduss/godrive/internal/file"
	"github.com/abduss/godrive/internal/server"
	"github.com/abduss/godrive/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := storage.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer dbPool.Close()

	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		log.Fatalf("connect minio: %v", err)
	}

	if err := storage.EnsureBucket(ctx, minioClient, cfg.MinIO.Bucket, cfg.MinIO.Region); err != nil {
		log.Fatalf("ensure bucket: %v", err)
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
		log.Printf("GoDrive API listening on %s", cfg.Server.Address())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("shutting down gracefully...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
