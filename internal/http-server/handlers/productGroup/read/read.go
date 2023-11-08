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
	Data *storage.ProductGroup `json:"data"`
}

type ProductReader interface {
	ReadProductGroup(product_group_id int64) (*storage.ProductGroup, error)
}

func New(log *slog.Logger, productReader ProductReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.productGroup.read.New"

		log = log.With(
			slog.String("op:", op),
			slog.String("request_id", middleware.GetReqID((r.Context()))),
		)

		product_group_id := chi.URLParam(r, "product_group_id")
		if product_group_id == "" {
			log.Info("product_group_id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		productGroup, err := strconv.ParseInt(product_group_id, 10, 32)
		if err != nil {
			log.Info("product_group_id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resProductGroup, err := productReader.ReadProductGroup(productGroup)
		if errors.Is(err, storage.ErrProductNotFound) {
			log.Info("product notfound", "product_id", product_group_id)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get product", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		jsonStr, err := json.Marshal(resProductGroup)
		if err != nil {
			log.Error("failed to get product json", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("got product", slog.String("productGroup", string(jsonStr)))

		responseOK(w, r, resProductGroup)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, productGroup *storage.ProductGroup) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Data:     productGroup,
	})
}
