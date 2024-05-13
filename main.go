package main

import (
	"context"
	"log"
	"restAPI/api"
	"restAPI/bootstrap"
	"restAPI/repository"
	"restAPI/service"
	"time"
)

func main() {
	cfg, err := bootstrap.NewConfig()
	if err != nil {
		log.Fatal("Problem with config load: ", err)
	}

	errorList := cfg.Validate()
	if errorList != nil {
		log.Fatal("Problem with config validation: ", errorList)
	}

	kafkaConn, err := bootstrap.KafkaConnect(cfg.KafkaAddr, cfg.KafkaTopic)
	if err != nil {
		log.Fatal("Problem with Kafka connection: ", err)
	}
	defer kafkaConn.Close()

	log.Println("kafka connection status: OK")

	db, err := bootstrap.DBConnect(cfg)
	if err != nil {
		log.Fatal("Problem with Postgres connection: ", err)
	}
	defer db.Close()

	log.Println("postgres DB connection status: OK")

	projRepo := repository.NewProjectRepository(db)
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	client, err := bootstrap.RedisConnect(cfg.RedisAddr)
	if err != nil {
		log.Fatal("Problem with Redis connection: ", err)
	}
	defer client.Close()

	cache := repository.NewRedisCache(userRepo, client)

	userServ := service.NewUserService(cache, authRepo, projRepo, kafkaConn)
	authServ := service.NewAuthService(authRepo, userRepo, kafkaConn)
	projServ := service.NewProjectRepository(projRepo, taskRepo, userRepo, kafkaConn)

	taskHandler := api.NewTaskHandler(projServ)
	projectHandler := api.NewProjectHandler(projServ)
	userHandler := api.NewUserHandler(userServ)
	authHandler := api.NewAuthHandler(authServ)

	mw := api.NewMiddleware(authServ)

	server := api.NewServer(taskHandler, projectHandler, userHandler, authHandler, cfg.HTTPPort, mw)

	go func() {
		for {
			err := userServ.SendVIPNotification(context.Background())
			if err != nil {
				log.Println(err)
			}

			time.Sleep(time.Minute)
		}
	}()

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
