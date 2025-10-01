package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListBrands() ([]Brand, error) {
	query := `SELECT * FROM brands`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Brand{}, err
	}
	defer rows.Close()

	var brands []Brand

	for rows.Next() {
		var brand Brand
		if err := rows.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.CreatedAt); err != nil {
			return []Brand{}, err
		}

		brands = append(brands, brand)
	}

	return brands, nil
}

func (s *Service) GetBrand(id string) (Brand, error) {
	var brand Brand

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return brand, err
	}

	query := "SELECT * FROM brands WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.CreatedAt)
	if err != nil {
		return brand, err
	}

	return brand, nil
}

func (s *Service) CreateBrand(brand Brand) error {
	query := "INSERT INTO brands (id,name,description) VALUES ($1,$2,$3);"
	validate := utils.NewValidate()

	err := validate.Struct(brand)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(context.Background(), query, id, brand.Name, brand.Description)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditBrand(id string, brand Brand) error {
	query := "UPDATE brands SET name=$1,description=$2 WHERE id=$3;"
	validate := utils.NewValidate()

	err := validate.Struct(brand)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(context.Background(), query, brand.Name, brand.Description, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteBrand(id string) error {
	query := "DELETE FROM brands WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
