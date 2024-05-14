package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"restAPI/entity"
)

type Error struct {
	Error string `json:"error"`
}

// sendError sending response with error based on error + status code.
func sendError(ctx context.Context, w http.ResponseWriter, err error) {
	l := entity.CtxLogger(ctx)

	l.Error("API error", "error", err)

	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, entity.ErrBadRequest):
		statusCode = http.StatusBadRequest
	case errors.Is(err, entity.ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, entity.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, entity.ErrForbidden):
		statusCode = http.StatusForbidden
	}

	w.WriteHeader(statusCode)

	err = json.NewEncoder(w).Encode(Error{Error: err.Error()})
	if err != nil {
		l.Error("API error", "error", err)
	}
}

func sendResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
