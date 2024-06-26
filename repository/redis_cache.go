package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"task-manager/entity"
	"time"
)

type RedisCache struct {
	client *redis.Client
	user   *UserRepository
}

func NewRedisCache(user *UserRepository, client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		user:   user,
	}
}

func (r *RedisCache) CreateUser(ctx context.Context, u entity.User) (entity.User, error) {
	l := entity.CtxLogger(ctx)

	user, err := r.user.CreateUser(ctx, u)
	if err != nil {
		return entity.User{}, err
	}

	key := fmt.Sprintf("user:%d", user.ID)

	value, err := json.Marshal(user)
	if err != nil {
		l.Error("redis error", "error", err)
		return user, nil
	}

	err = r.client.Set(ctx, key, string(value), time.Minute).Err()
	if err != nil {
		l.Error("redis error", "error", err)
		return user, nil
	}

	return user, nil
}

func (r *RedisCache) DeleteUser(ctx context.Context, id int64) error {
	return r.user.DeleteUser(ctx, id)
}

func (r *RedisCache) UserByID(ctx context.Context, id int64) (u entity.User, err error) {
	key := fmt.Sprintf("user:%d", id)

	result, err := r.client.Get(ctx, key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(result), &u)
		if err == nil {
			return u, nil
		}
	}

	return r.user.UserByID(ctx, id)
}

func (r *RedisCache) UsersToSendVIP(ctx context.Context) (users []entity.User, err error) {
	return r.user.UsersToSendVIP(ctx)
}

func (r *RedisCache) ProjectUsers(ctx context.Context, projectID int64) (users []entity.User, err error) {
	return r.user.ProjectUsers(ctx, projectID)
}

func (r *RedisCache) MarkNotification(ctx context.Context, email string, notification string) error {
	return r.user.MarkNotification(ctx, email, notification)
}
