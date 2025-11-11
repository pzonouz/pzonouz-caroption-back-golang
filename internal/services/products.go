package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) DeleteGeneratedProducts() ([]Product, error) {
	deleteGeneratedProductsQuery := `
		DELETE  FROM
		products
		WHERE generated = TRUE;
	`

	_, err := s.db.Exec(context.Background(), deleteGeneratedProductsQuery)
	if err != nil {
		return nil, err
	}

	return []Product{}, nil
}

func (s *Service) GenerateProducts() ([]Product, error) {
	ctx := context.Background()

	getGeneratableProductsQuery := `
		SELECT p.id, p.name, p.description, p.info, p.price, p.count,
		       p.category_id, p.brand_id, p.slug, p.keywords,
		       p.created_at, p.updated_at, p.image_id
		FROM products AS p
		WHERE p.generatable = TRUE AND p.generated = FALSE;
	`

	rows, err := s.db.Query(ctx, getGeneratableProductsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var generatableProducts []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Info, &p.Price, &p.Count,
			&p.CategoryID, &p.BrandID, &p.Slug, &p.Keywords,
			&p.CreatedAt, &p.UpdatedAt, &p.ImageID,
		); err != nil {
			return nil, err
		}

		generatableProducts = append(generatableProducts, p)
	}

	getGeneratorProductsQuery := `
	SELECT p.id, p.name, p.description, p.info, p.price, p.count,
	       p.category_id, p.brand_id, p.slug, p.keywords,
	       p.created_at, p.updated_at, p.image_id
	FROM products AS p
	LEFT JOIN categories AS c ON p.category_id = c.id
	LEFT JOIN categories AS parent ON c.parent_id = parent.id
	WHERE parent.generator = TRUE
	  AND p.generated = FALSE;
`

	rows, err = s.db.Query(ctx, getGeneratorProductsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var generatorProducts []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Info, &p.Price, &p.Count,
			&p.CategoryID, &p.BrandID, &p.Slug, &p.Keywords,
			&p.CreatedAt, &p.UpdatedAt, &p.ImageID,
		); err != nil {
			return nil, err
		}

		generatorProducts = append(generatorProducts, p)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	insertOrUpdateQuery := `
	INSERT INTO products (
		id, name, description, info, price, count,
		category_id, brand_id, slug, keywords,
		image_id, generated, generatable,show
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	ON CONFLICT (name)
	DO UPDATE SET
		description = EXCLUDED.description,
		info = EXCLUDED.info,
		price = EXCLUDED.price,
		count = EXCLUDED.count,
		category_id = EXCLUDED.category_id,
		brand_id = EXCLUDED.brand_id,
		slug = EXCLUDED.slug,
		keywords = EXCLUDED.keywords,
		image_id = EXCLUDED.image_id
	RETURNING id;
	`

	for _, generatorV := range generatorProducts {
		for _, generatableV := range generatableProducts {
			var newID uuid.UUID

			generatableV.Keywords = append(generatableV.Keywords, generatorV.Keywords...)

			price1, _ := strconv.Atoi(generatableV.Price.String)
			price2, _ := strconv.Atoi(generatorV.Price.String)
			totalPrice := price1 + price2

			// Insert or update and return the actual ID
			err = tx.QueryRow(ctx, insertOrUpdateQuery,
				uuid.New(), // tentative ID for insert
				generatableV.Name+" "+generatorV.Name,
				generatableV.Description.String+" "+generatorV.Description.String,
				generatableV.Info,
				strconv.Itoa(totalPrice),
				generatableV.Count.String,
				generatorV.CategoryID,
				generatableV.BrandID,
				fmt.Sprintf("%s_%s", generatableV.Slug.String, generatorV.Slug.String),
				generatableV.Keywords,
				generatorV.ImageID,
				true,  // generated
				false, // generatable
				true,
			).Scan(&newID)
			if err != nil {
				return nil, fmt.Errorf("insert or get product id failed: %v", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO images (id, name, image_url, product_id)
				SELECT gen_random_uuid(), name, image_url, $1
				FROM images WHERE product_id = $2
				ON CONFLICT (id) DO NOTHING;
			`, newID, generatableV.ID)
			if err != nil {
				return nil, fmt.Errorf(
					"copying images failed for product %s: %w",
					generatableV.ID,
					err,
				)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO product_parameter_values (
					id, product_id, parameter_id,
					bool_value, text_value, selectable_value
				)
				SELECT gen_random_uuid(), $1, parameter_id,
				       bool_value, text_value, selectable_value
				FROM product_parameter_values
				WHERE product_id = $2
				ON CONFLICT (product_id, parameter_id)
				DO UPDATE SET
					bool_value = EXCLUDED.bool_value,
					text_value = EXCLUDED.text_value,
					selectable_value = EXCLUDED.selectable_value;
			`, newID, generatableV.ID)
			if err != nil {
				return nil, fmt.Errorf(
					"copying product parameters failed for product %s: %w",
					generatableV.ID,
					err,
				)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return []Product{}, nil
}

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
		p.slug,
		p.keywords,
    p.created_at,
	  p.updated_at,
		p.generatable,
		p.generated,
    p.image_id,
    i.image_url,
		p.show,
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

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.CategoryID, &product.BrandID, &product.Slug, &product.Keywords, &product.CreatedAt, &product.UpdatedAt, &product.Generatable, &product.Generated, &product.ImageID, &product.ImageUrl, &product.Show, &product.ImageIDs, &product.Images, &productParameterValuesJSON); err != nil {
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

func (s *Service) RecentlyAddedProducts() ([]Product, error) {
	delayedTime := time.Now().AddDate(0, 0, -30)
	query := `
   	SELECT
    p.id,
    p.name,
    p.description,
    p.info,
    p.price,
    p.count,
		p.slug,
		p.keywords,
    p.created_at,
	  p.updated_at,
    p.image_id,
    i.image_url,
		p.show
FROM
    products p
LEFT JOIN images i ON p.image_id = i.id
WHERE p.created_at > $1
ORDER BY p.created_at ASC
`

	rows, err := s.db.Query(context.Background(), query, delayedTime)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product

		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.Slug, &product.Keywords, &product.CreatedAt, &product.UpdatedAt, &product.ImageID, &product.ImageUrl, &product.Show); err != nil {
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

	query := `
	WITH product_with_parent_category AS (
    SELECT
        p.id AS product_id,
        p.slug,
        c1.id AS category_id,
        c2.id AS parent_category_id
    FROM products p
    JOIN categories c1 ON p.category_id = c1.id
    LEFT JOIN categories c2 ON c1.parent_id = c2.id
    WHERE p.id = $1
)

SELECT
    p.id,
    p.name,
    p.description,
    p_agg.parameters,
    p.info,
    p.price,
    p.count,
    p.category_id,
    p.brand_id,
    b.name,
    p.slug,
    p.keywords,
    p.created_at,
    p.updated_at,
    p.generatable,
		p.generated,
    p.image_id,
    i.image_url,
    p.show,
    COALESCE(img_agg.image_ids, ARRAY[]::uuid[]) AS image_ids,
    COALESCE(img_agg.images, '[]'::json) AS images,
    COALESCE(ppv_agg.product_parameter_values, '[]'::json) AS product_parameter_values
FROM products p
JOIN product_with_parent_category pwpc ON p.id = pwpc.product_id
LEFT JOIN images i ON p.image_id = i.id
LEFT JOIN brands b ON p.brand_id = b.id

-- Aggregate images
LEFT JOIN (
    SELECT
        product_id,
        array_agg(id) AS image_ids,
        json_agg(json_build_object('id', id, 'imageUrl', image_url, 'name', name)) AS images
    FROM images
    WHERE product_id IS NOT NULL
    GROUP BY product_id
) img_agg ON img_agg.product_id = p.id

-- Aggregate product parameter values
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
) ppv_agg ON ppv_agg.product_id = p.id

-- Aggregate parameters from parameter groups linked to parent category
LEFT JOIN (
    SELECT
        pg.category_id,
        json_agg(json_build_object(
            'id', prm.id,
            'name', prm.name,
            'description', prm.description,
            'type', prm.type,
            'selectables', prm.selectables,
            'priority', prm.priority,
            'createdAt', prm.created_at
        ) ORDER BY prm.priority::int) AS parameters
    FROM parameter_groups pg
    JOIN parameters prm ON prm.parameter_group_id = pg.id
    GROUP BY pg.category_id
) p_agg ON p_agg.category_id = pwpc.parent_category_id
	WHERE p.id = $1;
	`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	var productParameterValuesJSON []byte

	err = row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Parameters,
		&product.Info,
		&product.Price,
		&product.Count,
		&product.CategoryID,
		&product.BrandID,
		&product.BrandName,
		&product.Slug,
		&product.Keywords,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Generatable,
		&product.Generated,
		&product.ImageID,
		&product.ImageUrl,
		&product.Show,
		&product.ImageIDs,
		&product.Images,
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

func (s *Service) GetProductBySlug(slug string) (Product, error) {
	var product Product

	query := `
	WITH product_with_parent_category AS (
    SELECT
        p.id AS product_id,
        p.slug,
        c1.id AS category_id,
        c2.id AS parent_category_id
    FROM products p
    JOIN categories c1 ON p.category_id = c1.id
    LEFT JOIN categories c2 ON c1.parent_id = c2.id
    WHERE p.slug = $1
)

SELECT
    p.id,
    p.name,
    p.description,
    p_agg.parameters,
    p.info,
    p.price,
    p.count,
    p.category_id,
    p.brand_id,
    b.name,
    p.slug,
    p.keywords,
    p.created_at,
    p.updated_at,
    p.generatable,
		p.generated,
    p.image_id,
    i.image_url,
    p.show,
    COALESCE(img_agg.image_ids, ARRAY[]::uuid[]) AS image_ids,
    COALESCE(img_agg.images, '[]'::json) AS images,
    COALESCE(ppv_agg.product_parameter_values, '[]'::json) AS product_parameter_values
FROM products p
JOIN product_with_parent_category pwpc ON p.id = pwpc.product_id
LEFT JOIN images i ON p.image_id = i.id
LEFT JOIN brands b ON p.brand_id = b.id

-- Aggregate images
LEFT JOIN (
    SELECT
        product_id,
        array_agg(id) AS image_ids,
        json_agg(json_build_object('id', id, 'imageUrl', image_url, 'name', name)) AS images
    FROM images
    WHERE product_id IS NOT NULL
    GROUP BY product_id
) img_agg ON img_agg.product_id = p.id

-- Aggregate product parameter values
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
) ppv_agg ON ppv_agg.product_id = p.id

-- Aggregate parameters from parameter groups linked to parent category
LEFT JOIN (
    SELECT
        pg.category_id,
        json_agg(json_build_object(
            'id', prm.id,
            'name', prm.name,
            'description', prm.description,
            'type', prm.type,
            'selectables', prm.selectables,
            'priority', prm.priority,
            'createdAt', prm.created_at
        ) ORDER BY prm.priority::int) AS parameters
    FROM parameter_groups pg
    JOIN parameters prm ON prm.parameter_group_id = pg.id
    GROUP BY pg.category_id
) p_agg ON p_agg.category_id = pwpc.parent_category_id
WHERE p.slug = $1;
	`

	row := s.db.QueryRow(context.Background(), query, slug)

	var productParameterValuesJSON []byte

	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Parameters,
		&product.Info,
		&product.Price,
		&product.Count,
		&product.CategoryID,
		&product.BrandID,
		&product.BrandName,
		&product.Slug,
		&product.Keywords,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Generatable,
		&product.Generated,
		&product.ImageID,
		&product.ImageUrl,
		&product.Show,
		&product.ImageIDs,
		&product.Images,
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
	query := "INSERT INTO products (id,name,description,info,price,count,category_id,brand_id,image_id,slug,keywords,generatable,show) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13);"
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
		product.Slug,
		product.Keywords,
		product.Generatable,
		product.Show,
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
	query := "UPDATE products SET name=$1,description=$2,info=$3,price=$4,count=$5,category_id=$6,brand_id=$7,image_id=$8,slug=$9,keywords=$10,generatable=$11,show=$12 WHERE id=$13;"
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
		product.Slug,
		product.Keywords,
		product.Generatable,
		product.Show,
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
		ppvId := uuid.New()

		_, err = tx.Exec(
			context.Background(),
			`
			INSERT INTO product_parameter_values (
					id,
					product_id,
					parameter_id,
					bool_value,
					text_value,
					selectable_value
			)
			VALUES (
					$1,  -- id
					$2,  -- product_id
					$3,  -- parameter_id
					$4,  -- bool_value
					$5,  -- text_value
					$6   -- selectable_value
			)
			ON CONFLICT (parameter_id, product_id)
			DO UPDATE SET
					bool_value = EXCLUDED.bool_value,
					text_value = EXCLUDED.text_value,
					selectable_value = EXCLUDED.selectable_value;
			`,
			ppvId,
			id,
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
		p.slug,
    p.created_at,
		p.updated_at,
		p.generatable,
		p.generated,
    p.image_id,
    i.image_url,
		p.show,
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
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Info, &product.Price, &product.Count, &product.CategoryID, &product.BrandID, &product.Slug, &product.CreatedAt, &product.UpdatedAt, &product.Generatable, &product.Generated, &product.ImageID, &product.ImageUrl, &product.Show, &product.ImageIDs, &product.Images); err != nil {
			return []Product{}, err
		}

		products = append(products, product)
	}

	return products, nil
}

func (s *Service) ProductsSearch(keywords string) ([]Product, error) {
	query := `
	SELECT
  p.id,
  p.name,
  p.info,
  p.price,
  p.count,
  p.category_id,
  p.brand_id,
  p.slug,
  p.created_at,
  p.updated_at,
  i.image_url
FROM
  products AS p
  LEFT JOIN images AS i ON p.image_id = i.id
WHERE
  p.show IS TRUE
  AND (
    p.fts @@ phraseto_tsquery('simple', normalize_persian($1))
    OR normalize_persian(p.name) ILIKE '%' || normalize_persian($1) || '%'
  )
ORDER BY
  p.updated_at DESC;
`

	rows, err := s.db.Query(context.Background(), query, keywords)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Info,
			&product.Price,
			&product.Count,
			&product.CategoryID,
			&product.BrandID,
			&product.Slug,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.ImageUrl,
			// &product.Rank, // <- optional if you have Rank field
		); err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, rows.Err()
}
