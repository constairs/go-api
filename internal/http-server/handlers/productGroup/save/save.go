package save

import (
	resp "go-api/internal/lib/api/response"
	"go-api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description,omitempty"`
}

type Response struct {
	resp.Response
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ProductGroupSaver interface {
	SaveProductGroup(title string, description string) (int64, error)
}

func New(log *slog.Logger, productGroupSaver ProductGroupSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.productGroup.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("faied to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		id, err := productGroupSaver.SaveProductGroup(req.Title, req.Description)
		if err != nil {
			log.Error("failed to add product group", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add product group"))

			return

		}

		log.Info("product group added", slog.Int64("id", id))

		idString := strconv.FormatInt(id, 10)

		responseOK(w, r, idString, req.Title, req.Description)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id string, title string, despription string) {
	render.JSON(w, r, Response{
		Response:    resp.OK(),
		Id:          id,
		Title:       title,
		Description: despription,
	})
}
