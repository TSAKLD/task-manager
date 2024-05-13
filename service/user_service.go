package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"restAPI/entity"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u entity.User) (entity.User, error)
	DeleteUser(ctx context.Context, id int64) error

	UserByID(ctx context.Context, id int64) (u entity.User, err error)
	UserByEmail(ctx context.Context, email string) (u entity.User, err error)
	Users(ctx context.Context) (users []entity.User, err error)
	ProjectUsers(ctx context.Context, projectID int64) (users []entity.User, err error)

	IsNotified(ctx context.Context, email string, notification string) error
	MarkNotification(ctx context.Context, email string, notification string) error
}

type UserService struct {
	kafka   *kafka.Conn
	user    UserRepository
	auth    AuthRepository
	project ProjectRepository
}

func NewUserService(user UserRepository, auth AuthRepository, project ProjectRepository, kafkaConn *kafka.Conn) *UserService {
	return &UserService{
		kafka:   kafkaConn,
		user:    user,
		auth:    auth,
		project: project,
	}
}

func (us *UserService) UserByID(ctx context.Context, id int64) (entity.User, error) {
	user, err := us.user.UserByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (us *UserService) DeleteUser(ctx context.Context, id int64) error {
	_, err := us.user.UserByID(ctx, id)
	if err != nil {
		return err
	}

	err = us.user.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) Users(ctx context.Context) ([]entity.User, error) {
	users, err := us.user.Users(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) ProjectUsers(ctx context.Context, projectID int64) ([]entity.User, error) {
	user := entity.AuthUser(ctx)

	projects, err := us.project.UserProjects(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	result := false

	for _, v := range projects {
		if projectID == v.ID {
			result = true
			break
		}
	}

	if !result {
		return nil, fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	users, err := us.user.ProjectUsers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) SendVIPNotification() {
	ctx := context.Background()
	ntf := entity.Notification{
		Subject:  "status update",
		Receiver: "",
		Message:  "You are VIP client now",
	}

	for {
		users, err := us.user.Users(ctx)
		if err != nil {
			log.Println(err)
		}

		for _, v := range users {
			if time.Now().Sub(v.CreatedAt).Hours() > 1 {
				ntf.Receiver = v.Email

				err := us.user.IsNotified(ctx, ntf.Receiver, ntf.Subject)
				if err != nil {
					if errors.Is(err, entity.ErrNotFound) {
						err = us.user.MarkNotification(ctx, ntf.Receiver, ntf.Subject)
						if err != nil {
							log.Println(err)
						}

						err = us.sendNotification(ctx, ntf)
						if err != nil {
							log.Println(err)
						}

						log.Println(fmt.Sprintf("%s: status update : %s", ntf.Receiver, ntf.Message))

						continue
					}

					log.Println(err)
				}
			}
		}

		time.Sleep(time.Minute)
	}
}

func (us *UserService) sendNotification(ctx context.Context, ntf entity.Notification) error {
	message := map[string]string{
		"subject":  fmt.Sprintf("New notification: %s", ntf.Subject),
		"receiver": ntf.Receiver,
		"message":  ntf.Message,
	}

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte("notification"),
		Value: b,
	}

	_, err = us.kafka.WriteMessages(msg)
	if err != nil {
		return err
	}

	return nil
}
