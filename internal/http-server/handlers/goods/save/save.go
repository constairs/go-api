package save

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "go-api/internal/lib/api/response"
	"go-api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	Title       string `json:"title" validate:"required"`
	Price       string `json:"price" validate:"required"`
	Description string `json:"description,omitempty"`
	ImgUrl      string `json:"imgUrl" validate:"required,url"`
	Weight      string `json:"weight" validate:"required"`
}

type Response struct {
	resp.Response
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
	ImgUrl      string  `json:"imgUrl"`
	Weight      int32   `json:"weight"`
}

type GoodsSaver interface {
	SaveGoods(title string, price float64, description string, imgUrl string, weight int32) (int64, error)
}

func New(log *slog.Logger, goodsSaver GoodsSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.goods.save.New"

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

		i32Weight := int32(weight)

		id, err := goodsSaver.SaveGoods(req.Title, priceFloat64, req.Description, req.ImgUrl, i32Weight)
		if err != nil {
			log.Error("failed to add goods", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add goods"))

			return
		}

		log.Info("goods added", slog.Int64("id", id))

		idString := strconv.FormatInt(id, 10)

		responseOK(w, r, idString, req.Title, priceFloat64, req.Description, req.ImgUrl, i32Weight)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id string, title string, price float64, description string, imgUrl string, weight int32) {
	render.JSON(w, r, Response{
		Response:    resp.OK(),
		Id:          id,
		Title:       title,
		Price:       price,
		Description: description,
		ImgUrl:      imgUrl,
		Weight:      weight,
	})
}
