package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/bcrypt"
	"task-manager/entity"
	"time"
)

type AuthRepository interface {
	UserByEmail(ctx context.Context, email string) (u entity.User, err error)
	CreateSession(ctx context.Context, sessionID uuid.UUID, userID int64, createdAt time.Time) error
	UserBySessionID(ctx context.Context, sessionID string) (u entity.User, err error)
	SaveVerificationCode(ctx context.Context, code string, userID int64) error
	VerifyUser(ctx context.Context, code string) error
}

type AuthService struct {
	auth  AuthRepository
	user  UserRepository
	kafka *kafka.Conn
}

func NewAuthService(auth AuthRepository, user UserRepository, kafkaConn *kafka.Conn) *AuthService {
	return &AuthService{
		auth:  auth,
		user:  user,
		kafka: kafkaConn,
	}
}

func (as *AuthService) RegisterUser(ctx context.Context, userTC entity.UserToCreate) (entity.User, error) {
	_, err := as.auth.UserByEmail(ctx, userTC.Email)
	if err == nil {
		return entity.User{}, fmt.Errorf("email %s already exist", userTC.Email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userTC.Password), 10)
	if err != nil {
		return entity.User{}, err
	}

	user := entity.User{
		Name:       userTC.Name,
		Password:   string(hash),
		Email:      userTC.Email,
		CreatedAt:  time.Now(),
		IsVerified: false,
	}

	user, err = as.user.CreateUser(ctx, user)
	if err != nil {
		return entity.User{}, err
	}

	user.Password = ""

	code := uuid.NewString()

	err = as.auth.SaveVerificationCode(ctx, code, user.ID)
	if err != nil {
		return entity.User{}, err
	}

	err = as.SendVerificationLink(ctx, code, user.Email)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (as *AuthService) Login(ctx context.Context, email string, password string) (uuid.UUID, error) {
	user, err := as.auth.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return uuid.UUID{}, entity.ErrUnauthorized
		}

		return uuid.UUID{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return uuid.UUID{}, entity.ErrUnauthorized
	}

	user.Password = ""

	if !user.IsVerified {
		return uuid.UUID{}, fmt.Errorf("%w: not verified, check your email", entity.ErrUnauthorized)
	}

	sessionID := uuid.New()

	createdAt := time.Now()

	err = as.auth.CreateSession(ctx, sessionID, user.ID, createdAt)
	if err != nil {
		return uuid.UUID{}, err
	}

	return sessionID, nil
}

func (as *AuthService) UserBySessionID(ctx context.Context, sessionID string) (entity.User, error) {
	return as.auth.UserBySessionID(ctx, sessionID)
}

func (as *AuthService) Verify(ctx context.Context, code string) error {
	return as.auth.VerifyUser(ctx, code)
}

func (as *AuthService) SendVerificationLink(_ context.Context, code string, email string) error {
	message := map[string]string{
		"subject":  "Verification",
		"receiver": email,
		"message":  fmt.Sprintf("Your Verification link is:http://localhost:8080/users/verify?code=%s", code),
	}

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(code),
		Value: b,
	}

	_, err = as.kafka.WriteMessages(msg)
	if err != nil {
		return err
	}

	return nil
}
