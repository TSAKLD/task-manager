package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"task-manager/entity"
	"time"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, t entity.Task) (entity.Task, error)
	TaskByID(ctx context.Context, id int64) (t entity.Task, err error)
	ProjectTasks(ctx context.Context, projectID int64) (tasks []entity.Task, err error)
	UserTasks(ctx context.Context, userID int64) (tasks []entity.Task, err error)
}

type ProjectRepository interface {
	CreateProject(ctx context.Context, project entity.Project) (entity.Project, error)
	UserProjects(ctx context.Context, userID int64) (projects []entity.Project, err error)
	ProjectByID(ctx context.Context, id int64) (p entity.Project, err error)
	DeleteProject(ctx context.Context, projectID int64) error
	AddProjectMember(ctx context.Context, code string) error

	SaveInvitationCode(ctx context.Context, code string, userID int64, projectID int64) error
}

type ProjectService struct {
	auth    AuthRepository
	project ProjectRepository
	user    UserRepository
	task    TaskRepository
	kafka   *kafka.Conn
}

func NewProjectRepository(auth AuthRepository, project ProjectRepository, task TaskRepository, user UserRepository, kafkaConn *kafka.Conn) *ProjectService {
	return &ProjectService{
		auth:    auth,
		project: project,
		user:    user,
		task:    task,
		kafka:   kafkaConn,
	}
}

func (ps *ProjectService) CreateProject(ctx context.Context, project entity.Project) (entity.Project, error) {
	user := entity.AuthUser(ctx)

	project.UserID = user.ID
	project.CreatedAt = time.Now()

	project, err := ps.project.CreateProject(ctx, project)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (ps *ProjectService) ProjectByID(ctx context.Context, id int64) (entity.Project, error) {
	user := entity.AuthUser(ctx)

	project, err := ps.project.ProjectByID(ctx, id)
	if err != nil {
		return entity.Project{}, err
	}

	if user.ID != project.UserID {
		return entity.Project{}, fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	return project, nil
}

func (ps *ProjectService) UserProjects(ctx context.Context) ([]entity.Project, error) {
	user := entity.AuthUser(ctx)
	return ps.project.UserProjects(ctx, user.ID)
}

func (ps *ProjectService) DeleteProject(ctx context.Context, projectID int64) error {
	user := entity.AuthUser(ctx)

	project, err := ps.project.ProjectByID(ctx, projectID)
	if err != nil {
		return err
	}

	if user.ID != project.UserID {
		return fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	err = ps.project.DeleteProject(ctx, projectID)
	if err != nil {
		return err
	}

	return nil
}

func (ps *ProjectService) CreateTask(ctx context.Context, cTask entity.TaskToCreate) (entity.Task, error) {
	project, err := ps.project.ProjectByID(ctx, cTask.ProjectID)
	if err != nil {
		return entity.Task{}, err
	}

	user := entity.AuthUser(ctx)

	if project.UserID != user.ID {
		return entity.Task{}, fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	task := entity.Task{
		Name:        cTask.Name,
		UserID:      user.ID,
		Description: cTask.Description,
		ProjectID:   cTask.ProjectID,
		CreatedAt:   time.Now(),
	}

	return ps.task.CreateTask(ctx, task)
}

func (ps *ProjectService) TaskByID(ctx context.Context, id int64) (entity.Task, error) {
	user := entity.AuthUser(ctx)

	task, err := ps.task.TaskByID(ctx, id)
	if err != nil {
		return entity.Task{}, err
	}

	if user.ID != task.UserID {
		return entity.Task{}, fmt.Errorf("%w: not your task", entity.ErrForbidden)
	}

	return task, nil
}

func (ps *ProjectService) ProjectTasks(ctx context.Context, projectID int64) ([]entity.Task, error) {
	project, err := ps.project.ProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	user := entity.AuthUser(ctx)

	if project.UserID != user.ID {
		return nil, fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	tasks, err := ps.task.ProjectTasks(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (ps *ProjectService) UserTasks(ctx context.Context) ([]entity.Task, error) {
	user := entity.AuthUser(ctx)

	tasks, err := ps.task.UserTasks(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (ps *ProjectService) AddProjectMember(ctx context.Context, code string) error {
	err := ps.project.AddProjectMember(ctx, code)
	if err != nil {
		return err
	}

	return nil
}

func (ps *ProjectService) InviteMemberRequest(ctx context.Context, projectID int64, email string) error {
	requester := entity.AuthUser(ctx)

	project, err := ps.project.ProjectByID(ctx, projectID)
	if err != nil {
		return err
	}

	if requester.ID != project.UserID {
		return fmt.Errorf("%w: not your project", entity.ErrForbidden)
	}

	user, err := ps.auth.UserByEmail(ctx, email)
	if err != nil {
		return err
	}

	code := uuid.NewString()

	err = ps.project.SaveInvitationCode(ctx, code, user.ID, projectID)
	if err != nil {
		return err
	}

	err = ps.SendInvite(ctx, email, code, project.Name)
	if err != nil {
		return err
	}

	return nil
}

func (ps *ProjectService) SendInvite(ctx context.Context, email string, code string, projectName string) error {
	message := map[string]string{
		"subject":  "Invitation",
		"receiver": email,
		"message":  fmt.Sprintf("You are invited to %s\nFollow link to accept invitation\n http://localhost:8080/projects/invite?code=%s", projectName, code),
	}

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(code),
		Value: b,
	}

	_, err = ps.kafka.WriteMessages(msg)
	if err != nil {
		return err
	}

	return nil
}
