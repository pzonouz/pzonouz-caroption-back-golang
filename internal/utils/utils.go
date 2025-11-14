package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[int(b)%len(charset)]
	}

	return string(bytes), nil
}

func HttpJsonFromArray[T any](array []T, w http.ResponseWriter) {
	val := reflect.ValueOf(array)

	var modArray []T

	var err error

	if val.Len() == 0 {
		modArray = make([]T, 0)

		err = json.NewEncoder(w).Encode(modArray)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(array)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func HttpJsonFromObject[T any](object T, w http.ResponseWriter) {
	var err error

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func ListFromQueryToResponse[T any](
	query func() ([]T, error),
	r *http.Request,
	w http.ResponseWriter,
) {
	objects, err := query()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	HttpJsonFromArray(objects, w)
}

func ListFromQueryToResponseById[T any](
	query func(string) ([]T, error),
	r *http.Request,
	w http.ResponseWriter,
	id string,
) {
	objects, err := query(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	HttpJsonFromArray(objects, w)
}

func ObjectFromQueryToResponse[T any](
	h func(string) (T, error),
	r *http.Request,
	w http.ResponseWriter,
	id string,
) {
	obj, err := h(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	HttpJsonFromObject(obj, w)
}

func DecodeBody[T any](r *http.Request, w http.ResponseWriter) (T, error) {
	var t T

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return t, err
	}

	return t, nil
}

func Uploader(w http.ResponseWriter, r *http.Request) error {
	// Limit upload size (e.g. 10MB)
	r.ParseMultipartForm(10 << 20) // 10 MB

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)

		return err
	}
	defer file.Close()

	randomString, err := RandomString(8)
	if err != nil {
		return err
	}
	// Create destination file
	ext := filepath.Ext(handler.Filename)
	fileName := fmt.Sprintf("./uploads/%s%s", randomString, ext)

	dst, err := os.Create(fileName)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)

		return err
	}
	defer dst.Close()

	// Copy uploaded content to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)

		return err
	}

	fmt.Fprint(w, strings.Split(fileName, "/")[2])

	return nil
}

func GetUserFromRequest(w http.ResponseWriter, r *http.Request) User {
	AuthHeader := r.Header.Get("Authorization")
	if AuthHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)

		return User{}
	}

	parts := strings.Split(AuthHeader, " ")
	if len(parts) != 2 {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)

		return User{}
	}

	tokenString := parts[1]

	claims := &AuthClaims{}

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

		return User{}
	}

	user := &User{
		ID:      claims.ID,
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
	}

	return *user
}

func ReplacePersianDigits(s string) string {
	persian := []rune("۰۱۲۳۴۵۶۷۸۹")

	english := []rune("0123456789")
	for i, p := range persian {
		s = strings.ReplaceAll(s, string(p), string(english[i]))
	}

	return s
}
