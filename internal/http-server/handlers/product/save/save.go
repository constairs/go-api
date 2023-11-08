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
	Price       string `json:"price" validate:"required"`
	Description string `json:"description,omitempty"`
	ImgUrl      string `json:"imgUrl" validate:"required,url"`
	Weight      string `json:"weight" validate:"required"`
	GroupId     string `json:"groupId" validate:"required"`
}

type Response struct {
	resp.Response
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	GroupId     string  `json:"group_id"`
	Description string  `json:"description,omitempty"`
	ImgUrl      string  `json:"img_url"`
	Weight      int32   `json:"weight"`
}

type ProductSaver interface {
	SaveProduct(title string, price float64, description string, imgUrl string, weight int64, group_id int64) (int64, error)
}

func New(log *slog.Logger, productSaver ProductSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.product.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

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

		priceFloat64, err := strconv.ParseFloat(req.Price, 64)
		if err != nil {
			log.Error("failed parse price value", sl.Err(err))

			render.JSON(w, r, resp.Error("failed parse price value"))

			return
		}

		weight, err := strconv.ParseInt(req.Weight, 10, 32)
		if err != nil {
			log.Error("failed parse weight value", sl.Err(err))

			render.JSON(w, r, resp.Error("failed parse weight value"))

			return
		}

		group_id, err := strconv.ParseInt(req.GroupId, 10, 32)
		if err != nil {
			log.Error("failed parse group id", sl.Err(err))

			render.JSON(w, r, resp.Error("failed parse group id"))

			return
		}

		id, err := productSaver.SaveProduct(req.Title, priceFloat64, req.Description, req.ImgUrl, weight, group_id)
		if err != nil {
			log.Error("failed to add product", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add product"))

			return
		}

		log.Info("product added", slog.Int64("id", id))

		idString := strconv.FormatInt(id, 10)

		i32Weight := int32(weight)

		groupIdString := strconv.FormatInt(group_id, 10)

		responseOK(w, r, idString, req.Title, priceFloat64, req.Description, req.ImgUrl, i32Weight, groupIdString)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id string, title string, price float64, description string, img_url string, weight int32, group_id string) {
	render.JSON(w, r, Response{
		Response:    resp.OK(),
		Id:          id,
		Title:       title,
		Price:       price,
		Description: description,
		ImgUrl:      img_url,
		Weight:      weight,
		GroupId:     group_id,
	})
}
