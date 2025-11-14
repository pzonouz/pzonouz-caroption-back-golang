package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/routes"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

type config struct {
	addr string
}

type application struct {
	config config
	db     *pgxpool.Pool
	mux    *chi.Mux
}

func (app *application) mount() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	service := services.New(app.db)
	routes.GenerateEntityRoutes(router, service)
	routes.GenerateProductRoutes(router, service)
	routes.GenerateCategoryRoutes(router, service)
	routes.GenerateBrandRoutes(router, service)
	routes.GenerateImageRoutes(router, service)
	routes.GenerateParameterGroupsRoutes(router, service)
	routes.GenerateParametersRoutes(router, service)
	routes.GenerateAuthRoutes(router, service)
	routes.GenerateArticleRoutes(router, service)
	router.Post("/upload-file", func(w http.ResponseWriter, r *http.Request) {
		_ = utils.Uploader(w, r)
	})

	return router
}

func (app *application) initDB() {
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_DBNAME")

	conn, err := pgxpool.New(
		context.Background(),
		"postgres://root:"+databasePassword+"@localhost:5432/"+databaseName,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	log.Print("Connected to Database")

	app.db = conn
}

func (app *application) run(mux *chi.Mux) error {
	server := &http.Server{
		Addr:    ":" + app.config.addr,
		Handler: mux,
	}
	log.Printf("Starting Server on %s", app.config.addr)

	return server.ListenAndServe()
}
