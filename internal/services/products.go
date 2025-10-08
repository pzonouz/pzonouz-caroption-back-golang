package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListProducts() ([]Product, error) {
	query := `
   	SELECT
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
    COALESCE(img_agg.image_ids, ARRAY[]::uuid[]) AS image_ids,
    COALESCE(img_agg.images, '[]'::json) AS images,
    COALESCE(ppv_agg.product_parameter_values, '[]'::json) AS product_parameter_values
FROM
    products p
LEFT JOIN images i ON p.image_id = i.id

-- Aggregate images separately
LEFT JOIN (
    SELECT
        product_id,
        array_agg(id) AS image_ids,
        json_agg(json_build_object('id', id, 'imageUrl', image_url, 'name', name)) AS images
    FROM images
    WHERE product_id IS NOT NULL
    GROUP BY product_id
) img_agg ON img_agg.product_id = p.id

-- Aggregate product parameter values separately
LEFT JOIN (
    SELECT
        product_id,
        json_agg(json_build_object(
            'id', id,
            'productId', product_id,
            'parameterId', parameter_id,
            'boolValue', bool_value,
            'textValue', text_value,
            'selectableValue', selectable_value,
            'createdAt', created_at
        )) AS product_parameter_values
    FROM product_parameter_values
    GROUP BY product_id
) ppv_agg ON ppv_agg.product_id = p.id;
`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product

		var productParameterValuesJSON []byte

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.CategoryID, &product.BrandID, &product.CreatedAt, &product.ImageID, &product.ImageUrl, &product.ImageIDs, &product.Images, &productParameterValuesJSON); err != nil {
			return []Product{}, err
		}

		if len(productParameterValuesJSON) > 0 {
			err = json.Unmarshal(productParameterValuesJSON, &product.ProductParameterValues)
			if err != nil {
				return nil, err
			}
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

	query := `
   	SELECT
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
    COALESCE(img_agg.image_ids, ARRAY[]::uuid[]) AS image_ids,
    COALESCE(img_agg.images, '[]'::json) AS images,
    COALESCE(ppv_agg.product_parameter_values, '[]'::json) AS product_parameter_values
FROM
    products p
LEFT JOIN images i ON p.image_id = i.id

-- Aggregate images separately
LEFT JOIN (
    SELECT
        product_id,
        array_agg(id) AS image_ids,
        json_agg(json_build_object('id', id, 'imageUrl', image_url, 'name', name)) AS images
    FROM images
    WHERE product_id IS NOT NULL
    GROUP BY product_id
) img_agg ON img_agg.product_id = p.id

-- Aggregate product parameter values separately
LEFT JOIN (
    SELECT
        product_id,
        json_agg(json_build_object(
            'id', id,
            'productId', product_id,
            'parameterId', parameter_id,
            'boolValue', bool_value,
            'textValue', text_value,
            'selectableValue', selectable_value,
            'createdAt', created_at
        )) AS product_parameter_values
    FROM product_parameter_values
    GROUP BY product_id
) ppv_agg ON ppv_agg.product_id = p.id WHERE p.id = $1;
	`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	var productParameterValuesJSON []byte

	err = row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Info,
		&product.Price,
		&product.Count,
		&product.CategoryID,
		&product.BrandID,
		&product.BrandName,
		&product.CreatedAt,
		&product.ImageID,
		&product.ImageUrl,
		&product.ImageIDs,
		&productParameterValuesJSON,
	)
	if len(productParameterValuesJSON) > 0 {
		_ = json.Unmarshal(productParameterValuesJSON, &product.ProductParameterValues)
	}

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

	for _, ppv := range product.ProductParameterValues {
		_, err = tx.Exec(
			context.Background(),
			`INSERT INTO product_parameter_values (product_id,parameter_id,text_value,bool_value,selectable_value) VALUES ($1,$2,$3,$4,$5)`,
			id,
			ppv.ParameterId,
			ppv.TextValue,
			ppv.BoolValue,
			ppv.SelectableValue,
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

	for _, ppv := range product.ProductParameterValues {
		_, err = tx.Exec(
			context.Background(),
			`UPDATE product_parameter_values SET product_id=$2,parameter_id=$3,bool_value=$4,text_value=$5,selectable_value=$6 WHERE id=$1`,
			ppv.ID,
			ppv.ProductID,
			ppv.ParameterId,
			ppv.BoolValue,
			ppv.TextValue,
			ppv.SelectableValue,
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

func (s *Service) ProductsInCategory(category_id string) ([]Product, error) {
	query := `
	SELECT
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
    COALESCE(
        array_agg(ims.id) FILTER (WHERE ims.id IS NOT NULL),
        ARRAY[]::uuid[]
    ) AS image_ids,
    COALESCE(
        json_agg(
	json_build_object('id', ims.id, 'imageUrl', ims.image_url,'name',ims.name)
        ) FILTER (WHERE ims.id IS NOT NULL),
        '[]'::json
    ) AS images
FROM
    products p
	  LEFT JOIN categories c ON p.category_id = c.id
    LEFT JOIN images i ON p.image_id = i.id
    LEFT JOIN images ims ON ims.product_id = p.id
WHERE c.parent_id = $1 OR c.id = $1
GROUP BY
    p.id, i.image_url;
`

	rows, err := s.db.Query(context.Background(), query, category_id)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.CategoryID, &product.BrandID, &product.CreatedAt, &product.ImageID, &product.ImageUrl, &product.ImageIDs, &product.Images); err != nil {
			return []Product{}, err
		}

		products = append(products, product)
	}

	return products, nil
}
