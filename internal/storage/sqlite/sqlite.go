package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	storage "go-api/internal/storage"
	"strconv"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createUrlTable, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url(
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createUrlTable.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createGroups, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS product_groups(
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    title TEXT NOT NULL,
		    description TEXT);
		CREATE INDEX IF NOT EXISTS idx_group_name ON product_groups(title);		
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+"createGroups", err)
	}

	createTestimonials, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS testimonials(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			author TEXT NOT NULL,
			author_img_url TEXT,
			title TEXT,
			text_content TEXT NOT NULL,
			img_content TEXT, 
			rating INTEGER NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createTestimonials", err)
	}

	createProducts, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS products(
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    title TEXT NOT NULL,
		    price REAL NOT NULL,
		    description TEXT,
		    img_url TEXT NOT NULL,
		    weight INTEGER NOT NULL,
		    group_id INTEGER NOT NULL,
		    quantity_per_serving INTEGER,
		    is_popular BOOL,
		    composition TEXT,
		    
		    FOREIGN KEY (group_id) REFERENCES product_groups (id) ON DELETE CASCADE);
			CREATE INDEX IF NOT EXISTS idx_product_name ON products(title);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProducts", err)
	}

	createProductGroups, err := db.Prepare(`
			CREATE TABLE IF NOT EXISTS group_product(
		    group_id INTEGER,
		    product_id INTEGER,
		    FOREIGN KEY (group_id) REFERENCES product_groups(id) ON DELETE CASCADE,
		    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProductGroups", err)
	}

	createProductPhotos, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS product_photo(
		    product_id INTEGER,
		    img_url TEXT NOT NULL,
		    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProductPhotos", err)
	}

	createProductTestimonials, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS product_testimonial(
		    product_id INTEGER,
		    testimonial_id INTEGER,
		    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
		    FOREIGN KEY (testimonial_id) REFERENCES testimonials(id) ON DELETE CASCADE);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProductTestimonials", err)
	}

	createProductRelated, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS product_related(
			product_Id INTEGER,
			related_id INTEGER NOT NULL,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
			FOREIGN KEY (related_id) REFERENCES products(id) ON DELETE CASCADE);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProductRelated", err)
	}

	createProductAdditives, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS product_additives(
		    product_Id INTEGER,
		    additive_id INTEGER NOT NULL,
		    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
		    FOREIGN KEY (additive_id) REFERENCES products(id) ON DELETE CASCADE);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op+".createProductAdditives", err)
	}

	_, err = createGroups.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProducts.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createTestimonials.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProductGroups.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createTestimonials.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProductPhotos.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProductTestimonials.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProductRelated.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = createProductAdditives.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) SaveProduct(title string, price float64, description string, img_url string, weight int64, group_id int64) (int64, error) {
	const op = "storage.sqlite.SaveProduct"

	stmt, err := s.db.Prepare("INSERT INTO products(title, price, description, img_url, weight, group_id) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(title, price, description, img_url, weight, group_id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ReadProduct(id int64) (*storage.Product, error) {
	const op = "storage.sqlite.ReadProduct"

	stmt, err := s.db.Prepare("SELECT title, price, description, img_url, weight, group_id, quantity_per_serving, is_popular, composition FROM products WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var title string
	var price float64
	var description string
	var img_url string
	var weight int64
	var group_id int64
	var quantity_per_serving any
	var is_popular any
	var composition any

	err = stmt.QueryRow(id).Scan(
		&title,
		&price,
		&description,
		&img_url,
		&weight,
		&group_id,
		&quantity_per_serving,
		&is_popular,
		&composition)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	id_string := strconv.FormatInt(id, 10)

	group_id_string := strconv.FormatInt(group_id, 10)

	product := &storage.Product{
		Id:          id_string,
		Title:       title,
		Price:       price,
		Description: description,
		ImgUrl:      img_url,
		Weight:      weight,
		GroupId:     group_id_string,
	}

	return product, nil
}

func (s *Storage) ReadProductList(limit int64, offset int64) (*[]storage.Product, int64, error) {
	const op = "storage.sqlite.ReadProductList"

	stmt, err := s.db.Prepare("SELECT id, title, price, description, img_url, weight, group_id, quantity_per_serving, is_popular, composition FROM products LIMIT ? OFFSET ?")
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var count int64

	err = s.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var arr []storage.Product

	for rows.Next() {
		var id int64
		var title string
		var price float64
		var description string
		var img_url string
		var weight int64
		var group_id int64
		var quantity_per_serving any
		var is_popular any
		var composition any

		err = rows.Scan(
			&id,
			&title,
			&price,
			&description,
			&img_url,
			&weight,
			&group_id,
			&quantity_per_serving,
			&is_popular,
			&composition)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, storage.ErrProductNotFound
		}
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		id_string := strconv.FormatInt(id, 10)

		group_id_string := strconv.FormatInt(group_id, 10)

		product := storage.Product{
			Id:          id_string,
			Title:       title,
			Price:       price,
			Description: description,
			ImgUrl:      img_url,
			Weight:      weight,
			GroupId:     group_id_string,
		}

		arr = append(arr, product)
	}

	return &arr, count, nil

}

func (s *Storage) SaveProductGroup(title string, description string) (int64, error) {
	const op = "storage.sqlite.SaveProductGroup"

	stmt, err := s.db.Prepare("INSERT INTO product_groups(title, description) VALUES (?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(title, description)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ReadProductGroup(product_group_id int64) (*storage.ProductGroup, error) {
	const op = "storage.sqlite.ReadProductGroup"

	stmt, err := s.db.Prepare("SELECT title, description FROM product_groups WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var title string
	var description string

	stmt.QueryRow(product_group_id).Scan(
		&title,
		&description,
	)

	product_group_id_string := strconv.FormatInt(product_group_id, 10)

	product_group := &storage.ProductGroup{
		Id:          product_group_id_string,
		Title:       title,
		Description: description,
	}

	return product_group, nil
}

func (s *Storage) ReadProductGroupList(limit int64, offset int64) (*[]storage.ProductGroup, int64, error) {
	const op = "storage.sqlite.ReadProductGroupList"

	stmt, err := s.db.Prepare("SELECT id, title, description FROM product_groups LIMIT ? OFFSET ?")
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var count int64

	err = s.db.QueryRow("SELECT COUNT(*) FROM product_groups").Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var arr []storage.ProductGroup

	for rows.Next() {
		var id int64
		var title string
		var description string

		err = rows.Scan(
			&id,
			&title,
			&description)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, storage.ErrProductNotFound
		}
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		id_string := strconv.FormatInt(id, 10)

		product_group := storage.ProductGroup{
			Id:          id_string,
			Title:       title,
			Description: description,
		}

		arr = append(arr, product_group)
	}

	return &arr, count, nil
}
