package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListImages() ([]Image, error) {
	query := "SELECT * FROM images"

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Image{}, err
	}
	defer rows.Close()

	var images []Image

	for rows.Next() {
		var image Image
		if err := rows.Scan(&image.ID, &image.Name, &image.ImageUrl, &image.ProductID, &image.CategoryID, &image.CreatedAt); err != nil {
			return []Image{}, err
		}

		images = append(images, image)
	}

	return images, nil
}

func (s *Service) GetImage(id string) (Image, error) {
	var image Image

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return image, err
	}

	query := "SELECT id,name,image_url,product_id,category_id,created_at FROM images WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(&image.ID, &image.Name, &image.ImageUrl, &image.ProductID, &image.CategoryID, &image.CreatedAt)
	if err != nil {
		return image, err
	}

	return image, nil
}

func (s *Service) CreateImage(image Image) error {
	query := "INSERT INTO images (id,name,image_url,product_id,category_id) VALUES($1,$2,$3,$4,$5)"
	validate := utils.NewValidate()

	err := validate.Struct(image)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(context.Background(), query, id, image.Name, image.ImageUrl, image.ProductID, image.CategoryID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditImage(id string, image Image) error {
	query := "UPDATE images SET name=$1,image_url=$2,product_id=$3,category_id=$4 WHERE id=$5;"
	validate := utils.NewValidate()

	err := validate.Struct(image)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(context.Background(), query, image.Name, image.ImageUrl, image.ProductID, image.CategoryID, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteImage(id string) error {
	query := "DELETE FROM images WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
