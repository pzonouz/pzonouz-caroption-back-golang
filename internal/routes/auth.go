package routes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

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
		router.Get("/reset_password/{email}", func(w http.ResponseWriter, r *http.Request) {
			email := chi.URLParam(r, "email")

			user, err := service.GetUser(email)
			if err != nil {
				http.Error(w, "No User", http.StatusNotFound)

				return
			}

			bytes := make([]byte, 16) // 16 bytes = 32 hex characters

			_, err = rand.Read(bytes)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)

				return
			}

			hexStr := hex.EncodeToString(bytes)
			user.Token.String = hexStr
			user.Token.Valid = true
			user.TokenExpires = time.Now().Add(time.Hour * 24)

			err = service.EditUser(user)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)

				return
			}

			var emailAddrs []string

			emailAddrs = append(emailAddrs, email)

			err = utils.SendMail(
				"info@caroptionshop.ir",
				emailAddrs,
				"Password Recovery",
				"peymanecu@gmail.com",
				"Peyman",
				"<div>Click this <a href='"+os.Getenv("BASE_URL")+"/reset-password-callback/"+hexStr+"'>Link</a> for Password Recovery,Expire Time:24 Hour</div>",
			)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)

				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		router.Post(
			"/reset_password_callback/{token}",
			func(w http.ResponseWriter, r *http.Request) {
				token := chi.URLParam(r, "token")

				user, err := service.GetUserByToken(token)
				if err != nil {
					http.Error(w, "Token Not Valid", http.StatusUnauthorized)

					return
				}

				diff := time.Until(user.TokenExpires)

				if diff < 0 {
					http.Error(w, "Token Not Valid", http.StatusUnauthorized)

					return
				}

				type Body struct {
					Password string `json:"password"`
				}

				var body Body

				decoder := json.NewDecoder(r.Body)

				err = decoder.Decode(&body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)

					return
				}

				err = service.SetUserPassword(user.ID, body.Password)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)

					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(""))
			},
		)
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
