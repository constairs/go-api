package storage

import "errors"

var (
	ErrURLNotFound     = errors.New("url not found")
	ErrURLExists       = errors.New("url exists")
	ErrProductNotFound = errors.New("product not found")
)

type Product struct {
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
	ImgUrl      string  `json:"imgUrl"`
	Weight      int64   `json:"weight"`
	GroupId     string  `json:"group_id"`
}

type ProductGroup struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
