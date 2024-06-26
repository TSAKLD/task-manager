package api

import (
	"context"
	"net/http"
	"strconv"
	"task-manager/entity"
)

type UserService interface {
	UserByID(ctx context.Context, id int64) (entity.User, error)
	ProjectUsers(ctx context.Context, projectID int64) ([]entity.User, error)

	DeleteUser(ctx context.Context, id int64) error
}

type UserHandler struct {
	user UserService
}

func NewUserHandler(user UserService) *UserHandler {
	return &UserHandler{user: user}
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qID := r.PathValue("id")

	id, err := strconv.ParseInt(qID, 10, 64)
	if err != nil {
		sendError(ctx, w, entity.ErrBadRequest)
		return
	}

	err = h.user.DeleteUser(ctx, id)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) UserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qID := r.PathValue("id")

	id, err := strconv.ParseInt(qID, 10, 64)
	if err != nil {
		sendError(ctx, w, entity.ErrBadRequest)
		return
	}

	user, err := h.user.UserByID(ctx, id)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	sendResponse(w, user)
}

func (h *UserHandler) ProjectUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qID := r.PathValue("project_id")

	projectID, err := strconv.ParseInt(qID, 10, 64)
	if err != nil {
		sendError(ctx, w, entity.ErrBadRequest)
		return
	}

	users, err := h.user.ProjectUsers(ctx, projectID)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	sendResponse(w, users)
}
