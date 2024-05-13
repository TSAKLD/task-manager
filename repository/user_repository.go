package repository

import (
	"context"
	"database/sql"
	"errors"
	"restAPI/entity"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, u entity.User) (entity.User, error) {
	q := "INSERT INTO users(name, password, email, created_at, is_verified) VALUES ($1, $2, $3, $4, $5) RETURNING id"

	err := r.db.QueryRowContext(ctx, q, u.Name, u.Password, u.Email, u.CreatedAt, u.IsVerified).Scan(&u.ID)
	if err != nil {
		return entity.User{}, err
	}

	return u, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	q := "DELETE FROM users WHERE id = $1"

	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UserByID(ctx context.Context, id int64) (u entity.User, err error) {
	q := "SELECT id, name, email, created_at, is_verified FROM users WHERE id = $1"

	err = r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.IsVerified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrNotFound
		}

		return u, err
	}

	return u, nil
}

func (r *UserRepository) UserByEmail(ctx context.Context, email string) (u entity.User, err error) {
	q := "SELECT id, name, email, created_at, is_verified FROM users WHERE email = $1"

	err = r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.IsVerified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrNotFound
		}

		return u, err
	}

	return u, nil
}

func (r *UserRepository) Users(ctx context.Context) (users []entity.User, err error) {
	q := "SELECT id, name, email, created_at, is_verified FROM users"

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User

		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.IsVerified)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) ProjectUsers(ctx context.Context, projectID int64) (users []entity.User, err error) {
	q := `SELECT u.id, u.name, u.email, u.created_at, u.is_verified
	FROM users u
	    JOIN projects_users pu ON pu.user_id = u.id
	WHERE pu.project_id = $1`

	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User

		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.IsVerified)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) IsNotified(ctx context.Context, email string, subject string) error {
	q := "SELECT email, subject FROM email_notifications WHERE email = $1 AND subject = $2"

	err := r.db.QueryRowContext(ctx, q, email, subject).Scan(&email, &subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *UserRepository) MarkNotification(ctx context.Context, email string, subject string) error {
	q := "INSERT INTO email_notifications (email, subject, created_at) VALUES($1, $2, $3)"

	_, err := r.db.ExecContext(ctx, q, email, subject, time.Now())
	if err != nil {
		return err
	}

	q = "UPDATE users SET vip_status = $1 WHERE email = $2"
	_, err = r.db.ExecContext(ctx, q, "active", email)
	if err != nil {
		return err
	}

	return nil
}
