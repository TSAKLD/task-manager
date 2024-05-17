package main

import (
	"context"
	"log/slog"
	"os"
	"task-manager/api"
	"task-manager/bootstrap"
	"task-manager/repository"
	"task-manager/service"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := bootstrap.NewConfig()
	if err != nil {
		logger.Error("Problem with config load: ", "error", err)
		return
	}

	errorList := cfg.Validate()
	if errorList != nil {
		logger.Error("Config validation error ", "error", err)
		return
	}

	kafkaConn, err := bootstrap.KafkaConnect(cfg.KafkaAddr, cfg.KafkaTopic)
	if err != nil {
		logger.Error("Problem with Kafka connection", "error", err)
	}
	defer kafkaConn.Close()

	logger.Info("kafka connection status: OK")

	db, err := bootstrap.DBConnect(cfg)
	if err != nil {
		logger.Error("Problem with Postgres connection", "error", err)
		return
	}
	defer db.Close()

	logger.Info("postgres DB connection status: OK")

	projRepo := repository.NewProjectRepository(db)
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	client, err := bootstrap.RedisConnect(cfg.RedisAddr)
	if err != nil {
		logger.Error("Problem with Redis connection", "error", err)
		return
	}
	defer client.Close()

	cache := repository.NewRedisCache(userRepo, client)

	userServ := service.NewUserService(cache, authRepo, projRepo, kafkaConn)
	authServ := service.NewAuthService(authRepo, userRepo, kafkaConn)
	projServ := service.NewProjectRepository(authRepo, projRepo, taskRepo, userRepo, kafkaConn)

	taskHandler := api.NewTaskHandler(projServ)
	projectHandler := api.NewProjectHandler(projServ)
	userHandler := api.NewUserHandler(userServ)
	authHandler := api.NewAuthHandler(authServ)

	mw := api.NewMiddleware(authServ, logger)

	server := api.NewServer(taskHandler, projectHandler, userHandler, authHandler, cfg.HTTPPort, mw)

	go func() {
		for {
			err := userServ.SendVIPNotification(context.Background(), *logger)
			if err != nil {
				logger.Error("Notification Sender error", "error", err)
			}

			time.Sleep(time.Minute)
		}
	}()

	err = server.Start()
	if err != nil {
		logger.Error("server start error", "error", err)
		return
	}
}
