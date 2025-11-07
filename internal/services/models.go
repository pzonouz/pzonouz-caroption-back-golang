package services

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Child struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name,omitempty"`
	ParentID    pgtype.UUID `json:"parentId"`
	ParentName  pgtype.Text `json:"parentName"`
	Description pgtype.Text `json:"description"`
	Priority    pgtype.Text `json:"priority"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type Category struct {
	ID          pgtype.UUID `json:"id"`
	Name        pgtype.Text `json:"name"`
	ParentID    pgtype.UUID `json:"parentId"`
	ParentName  pgtype.Text `json:"parentName"`
	Description pgtype.Text `json:"description"`
	Priority    pgtype.Text `json:"priority"`
	ImageID     pgtype.UUID `json:"imageId"`
	ImageUrl    pgtype.Text `json:"imageUrl"`
	Show        bool        `json:"show"`
	Children    []Child     `json:"children"`
	Slug        pgtype.Text `json:"slug"`
	Generator   pgtype.Bool `json:"generator"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Image struct {
	ID         pgtype.UUID `json:"id"`
	Name       string      `json:"name"`
	ImageUrl   string      `json:"imageUrl"`
	ProductID  pgtype.UUID `json:"productId"`
	CategoryID pgtype.UUID `json:"categoryId"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type Article struct {
	ID             pgtype.UUID   `json:"id"`
	Name           pgtype.Text   `json:"name"`
	Description    pgtype.Text   `json:"description"`
	ImageID        pgtype.UUID   `json:"imageId"`
	ImageUrl       pgtype.Text   `json:"imageUrl"`
	Slug           pgtype.Text   `json:"slug"`
	ShowInProducts bool          `json:"showInProducts"`
	Keywords       []pgtype.Text `json:"keywords"`
	CategoryID     pgtype.UUID   `json:"categoryId"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type Product struct {
	ID                     pgtype.UUID             `json:"id"`
	Name                   string                  `json:"name"`
	Description            pgtype.Text             `json:"description"`
	Info                   pgtype.Text             `json:"info"`
	Price                  pgtype.Text             `json:"price"`
	Count                  pgtype.Text             `json:"count"`
	CategoryID             pgtype.UUID             `json:"categoryId"`
	BrandID                pgtype.UUID             `json:"brandId"`
	BrandName              pgtype.Text             `json:"brandName"`
	Slug                   pgtype.Text             `json:"slug"`
	ImageID                pgtype.UUID             `json:"imageId"`
	ImageIDs               []pgtype.UUID           `json:"imageIds"`
	Images                 []any                   `json:"images"`
	ImageUrl               pgtype.Text             `json:"imageUrl"`
	Parameters             []Parameter             `json:"parameters"`
	ProductParameterValues []ProductParameterValue `json:"productParameterValues"`
	Generatable            pgtype.Bool             `json:"generatable"`
	Generated              pgtype.Bool             `json:"generated"`
	Keywords               []pgtype.Text           `json:"keywords"`
	Show                   pgtype.Bool             `json:"show"`
	Rank                   float64                 `json:"rank"`
	CreatedAt              time.Time               `json:"createdAt"`
	UpdatedAt              time.Time               `json:"updatedAt"`
}

type Brand struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name"`
	Description pgtype.Text `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type ParameterGroup struct {
	ID           pgtype.UUID `json:"id"`
	Name         string      `json:"name"`
	CategoryId   pgtype.UUID `json:"categoryId"`
	CategoryName pgtype.Text `json:"categoryName"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type Parameter struct {
	ID               pgtype.UUID   `json:"id"`
	Name             string        `json:"name"`
	Description      pgtype.Text   `json:"description"`
	Type             pgtype.Text   `json:"type"`
	ParameterGroupId pgtype.UUID   `json:"parameterGroupId"`
	Selectables      []pgtype.Text `json:"selectables"`
	Priority         pgtype.Text   `json:"priority"`
	CreatedAt        time.Time     `json:"createdAt"`
}

type ProductParameterValue struct {
	ID              pgtype.UUID `json:"id"`
	ProductID       pgtype.UUID `json:"productId"`
	ParameterId     pgtype.UUID `json:"parameterId"`
	TextValue       pgtype.Text `json:"textValue"`
	BoolValue       pgtype.Bool `json:"boolValue"`
	SelectableValue pgtype.Text `json:"selectableValue"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type User struct {
	ID        pgtype.UUID `json:"id"`
	Password  pgtype.Text `json:"password"`
	Email     pgtype.Text `json:"email"`
	IsAdmin   bool        `json:"isAdmin"`
	CreatedAt time.Time   `json:"createdAt"`
}
