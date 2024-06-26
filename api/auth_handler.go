package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"task-manager/entity"
	"time"
)

type AuthService interface {
	RegisterUser(ctx context.Context, userTC entity.UserToCreate) (entity.User, error)
	Login(ctx context.Context, email string, password string) (uuid.UUID, error)
	Verify(ctx context.Context, code string) error
	UserBySessionID(ctx context.Context, sessionID string) (entity.User, error)
	SendVerificationLink(ctx context.Context, code string, email string) error
}

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Registration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userToCreate entity.UserToCreate

	err := json.NewDecoder(r.Body).Decode(&userToCreate)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	err = userToCreate.Validate()
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	user, err := h.auth.RegisterUser(ctx, userToCreate)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	sendResponse(w, user)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user entity.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	sessionID, err := h.auth.Login(ctx, user.Email, user.Password)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		MaxAge:   24 * 60 * 60,
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	ctx := r.Context()

	err := h.auth.Verify(ctx, code)
	if err != nil {
		sendError(ctx, w, err)
		return
	}

	fmt.Fprint(w, "Verification Completed")
}
