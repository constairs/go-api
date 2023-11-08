package upload

import (
	"fmt"
	"go-api/internal/storage"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
)

type PhotoUploader interface {
	UploadPhoto(photoUrl string) (*storage.ProductGroup, error)
}

func New(log *slog.Logger, productReader PhotoUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.photo.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}

		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./photos/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		// fmt.Println("Метод:", r.Method)
		// if r.Method == "GET" {
		// 	crutime := time.Now().Unix()
		// 	h := md5.New()
		// 	io.WriteString(h, strconv.FormatInt(crutime, 10))
		// 	token := fmt.Sprintf("%x", h.Sum(nil))

		// 	t, _ := template.ParseFiles("upload.gtpl")
		// 	t.Execute(w, token)
		// } else {

		// }
	}
}
