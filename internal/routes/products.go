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
	mainRouter.Get("/recently_added_products", func(w http.ResponseWriter, r *http.Request) {
		utils.ListFromQueryToResponse(service.RecentlyAddedProducts, r, w)
	})
	mainRouter.Get("/product_by_slug/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		slug, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		utils.ObjectFromQueryToResponse(service.GetProductBySlug, r, w, slug)
	})
	mainRouter.Get("/products/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		keyword, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		utils.ObjectFromQueryToResponse(service.ProductsSearch, r, w, keyword)
	})
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/generate", func(router chi.Router) {
		router.Get("/products", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.GenerateProducts, r, w)
		})
		router.Get("/delete", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.DeleteGeneratedProducts, r, w)
		})
	})

	mainRouter.With(middlewares.AdminOrReadOnly).
		Route("/products_for_accounts", func(router chi.Router) {
			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				service.ListProductForAccountsWithSortFilterPagination(
					utils.DefaultInput(r.URL.Query().Get("sort"), ""),
					utils.DefaultInput(r.URL.Query().Get("sort_direction"), ""),
					r.URL.Query()["filter"],
					r.URL.Query()["filter_operand"],
					r.URL.Query()["filter_condition"],
					r.URL.Query().Get("count_in_page"),
					r.URL.Query().Get("offset"),
					w,
				)
			})
		})
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/products", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			service.ListProductsWithSortFilterPagination(
				utils.DefaultInput(r.URL.Query().Get("sort"), ""),
				utils.DefaultInput(r.URL.Query().Get("sort_direction"), ""),
				r.URL.Query()["filter"],
				r.URL.Query()["filter_operand"],
				r.URL.Query()["filter_condition"],
				r.URL.Query().Get("count_in_page"),
				r.URL.Query().Get("offset"),
				w,
			)
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
