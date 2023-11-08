package read

import (
	"encoding/json"
	"errors"
	resp "go-api/internal/lib/api/response"
	"go-api/internal/lib/logger/sl"
	"go-api/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Offset int64 `json:"offset" validate:"required"`
	Limit  int64 `json:"limit" validate:"required"`
}

type ListData struct {
	Items *[]storage.Product `json:"items"`
	Count int64              `json:"count"`
}

type Response struct {
	resp.Response
	Data ListData `json:"data"`
}

type ProductListReader interface {
	ReadProductList(count int64, offset int64) (*[]storage.Product, int64, error)
}

func New(log *slog.Logger, productListReader ProductListReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.product.read.New"

		log = log.With(
			slog.String("op:", op),
			slog.String("request_id", middleware.GetReqID((r.Context()))),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		resProducts, count, err := productListReader.ReadProductList(req.Limit, req.Offset)
		if errors.Is(err, storage.ErrProductNotFound) {
			log.Info("products not found")

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get products", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		jsonStr, err := json.Marshal(resProducts)
		if err != nil {
			log.Error("failed to get products json", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got products", slog.String("products", string(jsonStr)))
		log.Info("products count", slog.String("products", strconv.FormatInt(count, 10)))

		responseOK(w, r, resProducts, count)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, products *[]storage.Product, count int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Data:     ListData{Items: products, Count: count},
	})
}
