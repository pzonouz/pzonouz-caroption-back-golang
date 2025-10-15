package services

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) CreateUser(user User) error {
	query := `INSERT INTO users (id,email,password) VALUES ($1,$2,$3)`
	validate := utils.NewValidate()

	err := validate.Struct(user)
	if err != nil {
		return err
	}

	id := uuid.New()

	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password.String), 10)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		user.Email,
		cryptedPassword,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SignIn(user User) (string, error) {
	query := `SELECT id,email,password,is_admin FROM users  WHERE email = $1`
	validate := utils.NewValidate()

	err := validate.Struct(user)
	if err != nil {
		return "", err
	}

	var databaseUser User

	result := s.db.QueryRow(
		context.Background(),
		query,
		user.Email,
	)

	err = result.Scan(
		&databaseUser.ID,
		&databaseUser.Email,
		&databaseUser.Password,
		&databaseUser.IsAdmin,
	)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(databaseUser.Password.String),
		[]byte(user.Password.String),
	)
	if err != nil {
		return "", err
	}

	Claim := &utils.AuthClaims{
		ID:      databaseUser.ID.String(),
		Email:   databaseUser.Email.String,
		IsAdmin: databaseUser.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 30 * 12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, *Claim)

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) DeleteUser(id string) error {
	query := "DELETE FROM parameters WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
