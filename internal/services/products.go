package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListProducts() ([]Product, error) {
	query := `SELECT
    p.id,
    p.name,
    p.description,
    p.info,
    p.price,
    p.count,
    p.category_id,
    p.brand_id,
    p.created_at,
    p.image_id,
    i.image_url,
    COALESCE(array_agg(ims.id) FILTER (WHERE ims.id IS NOT NULL), ARRAY[]::uuid[]) AS image_ids
FROM
    products p
    LEFT JOIN images i ON p.image_id = i.id
    LEFT JOIN images ims ON ims.product_id = p.id
GROUP BY
    p.id, i.image_url;
`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.CategoryID, &product.BrandID, &product.CreatedAt, &product.ImageID, &product.ImageUrl, &product.ImageIDs); err != nil {
			return []Product{}, err
		}

		products = append(products, product)
	}

	return products, nil
}

func (s *Service) GetProduct(id string) (Product, error) {
	var product Product

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return product, err
	}

	query := "SELECT p.id,p.name,p.description,p.info,p.price,p.count,p.category_id,p.brand_id,p.created_at,p.image_id,i.image_url FROM products as p LEFT JOIN images as i ON p.image_id=i.id WHERE p.id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Info,
		&product.Price,
		&product.Count,
		&product.CategoryID,
		&product.BrandID,
		&product.CreatedAt,
		&product.ImageID,
		&product.ImageUrl,
	)
	if err != nil {
		return product, err
	}

	return product, nil
}

func (s *Service) CreateProduct(product Product) error {
	query := "INSERT INTO products (id,name,description,info,price,count,category_id,brand_id,image_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9);"
	validate := utils.NewValidate()

	err := validate.Struct(product)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	id := uuid.New()

	_, err = tx.Exec(
		context.Background(),
		query,
		id,
		product.Name,
		product.Description,
		product.Info,
		product.Price,
		product.Count,
		product.CategoryID,
		product.BrandID,
		product.ImageID,
	)
	if err != nil {
		return err
	}

	for _, imgID := range product.ImageIDs {
		_, err = tx.Exec(context.Background(),
			`UPDATE images SET product_id=$1 WHERE id=$2`,
			id, imgID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (s *Service) EditProduct(id string, product Product) error {
	query := "UPDATE products SET name=$1,description=$2,info=$3,price=$4,count=$5,category_id=$6,brand_id=$7,image_id=$8 WHERE id=$9;"
	validate := utils.NewValidate()

	err := validate.Struct(product)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = s.db.Exec(
		context.Background(),
		query,
		product.Name,
		product.Description,
		product.Info,
		product.Price,
		product.Count,
		product.CategoryID,
		product.BrandID,
		product.ImageID,
		id,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(),
		`UPDATE images SET product_id = NULL WHERE product_id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	for _, imgID := range product.ImageIDs {
		_, err = tx.Exec(context.Background(),
			`UPDATE images SET product_id=$1 WHERE id=$2`,
			id, imgID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (s *Service) DeleteProduct(id string) error {
	query := "DELETE FROM products WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
