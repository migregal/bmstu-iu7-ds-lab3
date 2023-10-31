package v1

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/httpwrapper"
	"github.com/migregal/bmstu-iu7-ds-lab2/rating/core/ports/ratings"
)

type Core interface {
	GetUserRating(context.Context, string) (ratings.Rating, error)
	UpdateUserRating(context.Context, string, int) error
}

type api struct {
	lg   *slog.Logger
	core Core
}

func InitListener(mx *echo.Echo, lg *slog.Logger, core Core) error {
	gr := mx.Group("/api/v1")

	a := api{lg: lg, core: core}

	gr.GET("/rating", httpwrapper.WrapRequest(lg, a.GetRating))
	gr.PATCH("/rating", httpwrapper.WrapRequest(lg, a.UpdateRating))

	return nil
}
