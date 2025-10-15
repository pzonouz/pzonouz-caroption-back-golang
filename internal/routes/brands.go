package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateBrandRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/brands", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResonse(service.ListBrands, r, w)
		})

		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetBrand, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			brand, err := utils.DecodeBody[services.Brand](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateBrand(brand)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			brand, err := utils.DecodeBody[services.Brand](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditBrand(id, brand)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
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
