package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lohanguedes/first-api/internal/store"
)

type userRepo interface {
	FindAll() []store.User
	FindById(uuid string) (store.User, error)
	Insert(firstName, lastName, bio string) (store.User, error)
	Update(uuid string, u store.User) (store.User, error)
	Delete(uuid string) (store.User, error)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"health": "ok",
		})
	})

	app := application{
		db:     store.UserRepo{},
		router: r,
	}
	app.bindRoutes()

	fmt.Println("start listening on port :3000")
	http.ListenAndServe("localhost:3000", app.router)
}
