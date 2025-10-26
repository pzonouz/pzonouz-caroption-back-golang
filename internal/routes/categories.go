package routes

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func GenerateCategoryRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Get("/parent_categories", func(w http.ResponseWriter, r *http.Request) {
		utils.ListFromQueryToResonse(service.ListParentCategories, r, w)
	})
	mainRouter.Get("/category_by_slug/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		slug, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		utils.ObjectFromQueryToResponse(service.GetCategoryBySlug, r, w, slug)
	})
	mainRouter.Get("/products_in_category/{id}", func(w http.ResponseWriter, r *http.Request) {
		stringId := chi.URLParam(r, "id")
		utils.ListFromQueryToResonseById(
			service.ProductsInCategory,
			r,
			w,
			stringId,
		)
	})

	// mainRouter.With(middlewares.AdminOrReadOnly).Route("/categories", func(router chi.Router) {
	mainRouter.Route("/categories", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResonse(service.ListCategories, r, w)
		})

		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetCategory, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			category, err := utils.DecodeBody[services.Category](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateCategory(category)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})

		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			category, err := utils.DecodeBody[services.Category](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditCategory(id, category)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteCategory(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
