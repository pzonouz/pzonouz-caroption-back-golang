package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

type Token struct {
	Access string `json:"access"`
}

type UserData struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

func GenerateAuthRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Route("/auth", func(router chi.Router) {
		router.Post("/signin", func(w http.ResponseWriter, r *http.Request) {
			user, err := utils.DecodeBody[services.User](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			token, err := service.SignIn(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)

				return
			}

			w.Header().Set("Content-Type", "application/json")

			tokenData := &Token{Access: token}
			encoder := json.NewEncoder(w)

			err = encoder.Encode(tokenData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		router.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			user, err := utils.DecodeBody[services.User](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateUser(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			AuthHeader := r.Header.Get("Authorization")
			if AuthHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)

				return
			}

			parts := strings.Split(AuthHeader, " ")
			if len(parts) != 2 {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)

				return
			}

			tokenString := parts[1]

			claims := &utils.AuthClaims{}

			token, err := jwt.ParseWithClaims(
				tokenString,
				claims,
				func(token *jwt.Token) (any, error) {
					return []byte(os.Getenv("SECRET")), nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			)
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)

				return
			}

			userData := &UserData{
				ID:      claims.ID,
				Email:   claims.Email,
				IsAdmin: claims.IsAdmin,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(userData)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteBrand(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
