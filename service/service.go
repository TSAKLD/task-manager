package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"restAPI/entity"
	"restAPI/repository"
	"time"
)

type UserService struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *UserService {
	return &UserService{repo: repo}
}

// user manipulations

func (us *UserService) RegisterUser(user entity.User) (entity.User, error) {
	_, err := us.repo.UserByEmail(user.Email)
	if err == nil {
		return entity.User{}, errors.New("email %v already exist")
	}

	user.CreatedAt = time.Now()

	user, err = us.repo.CreateUser(user)
	if err != nil {
		return entity.User{}, err
	}

	user.Password = ""

	return user, nil
}

func (us *UserService) UserByID(id int64) (entity.User, error) {
	user, err := us.repo.UserByID(id)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (us *UserService) DeleteUser(id int64) error {
	_, err := us.repo.UserByID(id)
	if err != nil {
		return err
	}

	err = us.repo.DeleteUser(id)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) Users() ([]entity.User, error) {
	users, err := us.repo.Users()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) Login(email string, password string) (uuid.UUID, error) {
	user, err := us.repo.UserByEmailAndPassword(email, password)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return uuid.UUID{}, entity.ErrUnauthorized
		}

		return uuid.UUID{}, err
	}

	sessionID := uuid.New()

	createdAt := time.Now()

	err = us.repo.CreateSession(sessionID, user.ID, createdAt)
	if err != nil {
		return uuid.UUID{}, err
	}

	return sessionID, nil
}

func (us *UserService) UserBySessionID(sessionID string) (entity.User, error) {
	return entity.User{}, nil
}

// project manipulations

func (us *UserService) CreateProject(project entity.Project) (entity.Project, error) {
	project, err := us.repo.CreateProject(project)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (us *UserService) ProjectByID(id int64) (entity.Project, error) {
	project, err := us.repo.ProjectByID(id)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (us *UserService) Projects(ownerID int64) ([]entity.Project, error) {
	return us.repo.UserProjects(ownerID)
}

func (us *UserService) DeleteProject(ownerID int64, projectID int64) error {
	project, err := us.repo.ProjectByID(projectID)
	if err != nil {
		return err
	}

	if ownerID != project.UserID {
		return fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	err = us.repo.DeleteProject(projectID)
	if err != nil {
		return err
	}

	return nil
}
