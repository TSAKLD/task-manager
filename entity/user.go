package entity

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Password   string    `json:"password,omitempty"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	IsVerified bool      `json:"is_verified"`
}

func AuthUser(ctx context.Context) User {
	return ctx.Value("user").(User)
}

type UserToCreate struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (utc *UserToCreate) Validate() error {
	if utc.Name == "" {
		return errors.New("invalid name field")
	}

	if utc.Password == "" {
		return errors.New("invalid password field")
	}

	if utc.Email == "" {
		return errors.New("invalid email field")
	}

	return nil
}
