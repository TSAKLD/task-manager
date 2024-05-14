package entity

import (
	"context"
	"log/slog"
)

func CtxLogger(ctx context.Context) *slog.Logger {
	return ctx.Value("logger").(*slog.Logger)
}
