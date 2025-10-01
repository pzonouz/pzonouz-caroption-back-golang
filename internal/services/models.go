package services

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Child struct {
	ID        pgtype.UUID `json:"id"`
	Name      string      `json:"name,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
}

type Category struct {
	ID          pgtype.UUID `json:"id"`
	Name        pgtype.Text `json:"name"        validate:"required,notblank"`
	ParentID    pgtype.UUID `json:"parentId"`
	ParentName  pgtype.Text `json:"parentName"`
	Description pgtype.Text `json:"description"`
	Prioirity   pgtype.Text `json:"prioirity"`
	Children    []Child     `json:"children"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type Image struct {
	ID         pgtype.UUID `json:"id"`
	Name       string      `json:"name"`
	ImageUrl   string      `json:"imageUrl"`
	ProductID  pgtype.UUID `json:"productId"`
	CategoryID pgtype.UUID `json:"categoryId"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type Product struct {
	ID          pgtype.UUID   `json:"id"`
	Name        string        `json:"name"`
	Description pgtype.Text   `json:"description"`
	Info        pgtype.Text   `json:"info"`
	Price       pgtype.Text   `json:"price"`
	Count       pgtype.Text   `json:"count"`
	CategoryID  pgtype.UUID   `json:"categoryId"`
	BrandID     pgtype.UUID   `json:"brandId"`
	ImageID     pgtype.UUID   `json:"imageId"`
	ImageIDs    []pgtype.UUID `json:"imageIds"`
	Images      []Image       `json:"images"`
	ImageUrl    pgtype.Text   `json:"imageUrl"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type Brand struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name"`
	Description pgtype.Text `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
}
