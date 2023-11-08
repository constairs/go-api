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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Data *storage.Product `json:"data"`
}

type ProductReader interface {
	ReadProduct(id int64) (*storage.Product, error)
}

func New(log *slog.Logger, productReader ProductReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.product.read.New"

		log = log.With(
			slog.String("op:", op),
			slog.String("request_id", middleware.GetReqID((r.Context()))),
		)

		product_id := chi.URLParam(r, "product_id")
		if product_id == "" {
			log.Info("product_id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		product, err := strconv.ParseInt(product_id, 10, 32)
		if err != nil {
			log.Info("product_id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resProduct, err := productReader.ReadProduct(product)
		if errors.Is(err, storage.ErrProductNotFound) {
			log.Info("product notfound", "product_id", product_id)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get product", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		jsonStr, err := json.Marshal(resProduct)
		if err != nil {
			log.Error("failed to get product json", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got product", slog.String("product", string(jsonStr)))

		responseOK(w, r, resProduct)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, product *storage.Product) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Data:     product,
	})
}
