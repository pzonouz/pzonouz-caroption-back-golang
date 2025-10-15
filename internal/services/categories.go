package services

import (
	"context"
	"log"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParentCategories() ([]Category, error) {
	query := `SELECT 
    p.id,
    p.name,
    p.parent_id,
		p.description,
		p.prioirity,
		i.image_url,
		p.image_id,
    p.created_at,
	  p.updated_at,	
    COALESCE(
        json_agg(
            json_build_object(
                'id', c.id,
                'name', c.name,
                'createdAt', c.created_at
            )
        ) FILTER (WHERE c.id IS NOT NULL),
        '[]'
    ) AS children
FROM categories p
LEFT JOIN categories c ON c.parent_id = p.id
LEFT JOIN images i ON p.image_id = i.id
WHERE p.parent_id IS NULL
GROUP BY p.id, p.name, p.parent_id, p.created_at,i.image_url ORDER BY p.prioirity;`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Category{}, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.Description, &category.Prioirity, &category.ImageUrl, &category.ImageID, &category.CreatedAt, &category.UpdatedAt, &category.Children); err != nil {
			return []Category{}, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return []Category{}, err
	}

	return categories, nil
}

func (s *Service) ListCategories() ([]Category, error) {
	query := `SELECT
    c.id,
    c.name,
    c.parent_id,
    p.name AS parent_name,
    c.description,
    c.prioirity,
		c.image_id,
		i.image_url,
    c.created_at,
	  c.updated_at
FROM
    categories AS c
    LEFT JOIN categories p ON c.parent_id = p.id
    LEFT JOIN images i ON c.image_id = i.id
GROUP BY
    c.id,
		i.image_url,
    p.name;
	`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Category{}, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.ParentName, &category.Description, &category.Prioirity, &category.ImageID, &category.ImageUrl, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return []Category{}, err
		}

		category.Children = []Child{}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Service) GetCategory(id string) (Category, error) {
	var category Category

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return category, err
	}

	query := "SELECT id,name,parent_id,description,prioirity,image_id,created_at,updated_at FROM categories WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&category.ID,
		&category.Name,
		&category.ParentID,
		&category.Description,
		&category.Prioirity,
		&category.ImageID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *Service) CreateCategory(category Category) error {
	query := "INSERT INTO categories (id,name,parent_id,description,prioirity,image_id) VALUES ($1,$2,$3,$4,$5,$6);"
	validate := utils.NewValidate()

	err := validate.Struct(category)
	if err != nil {
		return err
	}

	id := uuid.New()
	log.Print(id)

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		category.Name,
		category.ParentID,
		category.Description,
		category.Prioirity,
		category.ImageID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditCategory(id string, category Category) error {
	query := "UPDATE categories SET name=$1,parent_id=$2,image_id=$3,prioirity=$4 WHERE id=$5;"
	validate := utils.NewValidate()

	err := validate.Struct(category)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		category.Name,
		category.ParentID,
		category.ImageID,
		category.Prioirity,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCategory(id string) error {
	query := "DELETE FROM categories WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
