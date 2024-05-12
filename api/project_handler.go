package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"restAPI/entity"
	"strconv"
)

type ProjectService interface {
	CreateProject(ctx context.Context, project entity.Project) (entity.Project, error)
	DeleteProject(ctx context.Context, projectID int64) error

	ProjectByID(ctx context.Context, id int64) (entity.Project, error)
	UserProjects(ctx context.Context) ([]entity.Project, error)

	AddProjectMember(ctx context.Context, code string) error
	InviteMemberRequest(ctx context.Context, projectID int64, email string) error
	SendInvite(ctx context.Context, email string, code string, projectName string) error
}

type ProjectHandler struct {
	project ProjectService
}

func NewProjectHandler(project ProjectService) *ProjectHandler {
	return &ProjectHandler{project: project}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var project entity.Project

	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		sendError(w, err)
		return
	}

	project, err = h.project.CreateProject(ctx, project)
	if err != nil {
		sendError(w, err)
		return
	}

	sendResponse(w, project)
}

func (h *ProjectHandler) UserProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.project.UserProjects(ctx)
	if err != nil {
		sendError(w, err)
		return
	}

	sendResponse(w, projects)
}

func (h *ProjectHandler) ProjectByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qID := r.PathValue("id")
	projectID, err := strconv.ParseInt(qID, 10, 64)
	if err != nil {
		sendError(w, errors.New("'id' must be an integer"))
		return
	}

	project, err := h.project.ProjectByID(ctx, projectID)
	if err != nil {
		sendError(w, err)
		return
	}

	sendResponse(w, project)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qID := r.PathValue("id")
	projectID, err := strconv.ParseInt(qID, 10, 64)
	if err != nil {
		sendError(w, errors.New("'id' must be an integer"))
		return
	}

	err = h.project.DeleteProject(ctx, projectID)
	if err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type InviteMemberRequest struct {
	ProjectID int64  `json:"project_id"`
	Email     string `json:"email"`
}

func (h *ProjectHandler) AcceptProjectInvitation(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	ctx := r.Context()

	err := h.project.AddProjectMember(ctx, code)
	if err != nil {
		sendError(w, err)
		return
	}

	fmt.Fprint(w, "Invitation Accepted")
}

func (h *ProjectHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	var request InviteMemberRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, err)
		return
	}

	ctx := r.Context()

	err = h.project.InviteMemberRequest(ctx, request.ProjectID, request.Email)
	if err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
