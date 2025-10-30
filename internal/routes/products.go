package routes

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateProductRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Get("/generate_products", func(w http.ResponseWriter, r *http.Request) {
		utils.ListFromQueryToResonse(service.GenerateProducts, r, w)
	})
	mainRouter.Get("/product_by_slug/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		slug, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		utils.ObjectFromQueryToResponse(service.GetProductBySlug, r, w, slug)
	})
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/products", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResonse(service.ListProducts, r, w)
		})

		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetProduct, r, w, id)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			product, err := utils.DecodeBody[services.Product](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateProduct(product)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			product, err := utils.DecodeBody[services.Product](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditProduct(id, product)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteProduct(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
