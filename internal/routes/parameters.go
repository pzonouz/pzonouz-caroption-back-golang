package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func GenerateParametersRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Route("/parameters", func(router chi.Router) {
		router.Get("/by-group/{id}", func(w http.ResponseWriter, r *http.Request) {
			categoryId := chi.URLParam(r, "id")
			utils.ListFromQueryToResonseById(service.ListParametersByCategory, r, w, categoryId)
		})
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResonse(service.ListParameters, r, w)
		})
		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetParameter, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			parameter, err := utils.DecodeBody[services.Parameter](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateParameter(parameter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			parameter, err := utils.DecodeBody[services.Parameter](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditParameter(id, parameter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteParameter(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
