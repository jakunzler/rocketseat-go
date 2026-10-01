package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lohanguedes/first-api/internal/store"
	"github.com/lohanguedes/first-api/internal/validator"
)

type application struct {
	db     userRepo
	router *chi.Mux
}

type UserRequest struct {
	// Only used on server-side NOT client-expected
	Validator validator.Validator `json:"-"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

func jsonResponse[K comparable, V any](w http.ResponseWriter, status int, response map[K]V) {
	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unexpected internal server error. try again later"))
		return
	}
	w.WriteHeader(status)
	w.Write(resp)
}

func (app *application) addUser(w http.ResponseWriter, r *http.Request) {
	var body UserRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		jsonResponse(w, http.StatusUnprocessableEntity, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Validate user input
	body.Validator.CheckField(validator.NotBlank(body.FirstName), "first_name", "this field cannot be blank")
	body.Validator.CheckField(validator.NotBlank(body.LastName), "last_name", "this field cannot be blank")
	body.Validator.CheckField(validator.NotBlank(body.Bio), "bio", "this field cannot be blank")

	body.Validator.CheckField(validator.MinChars(body.FirstName, 2) && validator.MaxChars(body.FirstName, 20), "first_name", "first_name must be at least 2 chars long and smaller than 20 chars")
	body.Validator.CheckField(validator.MinChars(body.LastName, 2) && validator.MaxChars(body.LastName, 20), "last_name", "last_name must be at least 2 chars long and smaller than 20 chars")
	body.Validator.CheckField(validator.MinChars(body.Bio, 20) && validator.MaxChars(body.LastName, 450), "bio", "bio must be at least 20 chars long and smaller than 450 chars")

	if !body.Validator.Valid() {
		jsonResponse(w, http.StatusBadRequest, body.Validator.FieldErrors)
		return
	}

	user, err := app.db.Insert(body.FirstName, body.LastName, body.Bio)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{
			"error": "could not insert user on database",
		})
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]any{
		"user": user,
	})
}

func (app *application) getAllUsers(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]any{
		"users": app.db.FindAll(),
	})
}

func (app *application) getUserById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := uuid.Parse(id)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{
			"message": "id must be of type uuid",
			"error":   err.Error(),
		})
		return
	}

	user, err := app.db.FindById(id)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{
			"message": fmt.Sprintf("user with id: %s not found", id),
		})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (app *application) deleteUserById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := uuid.Parse(id)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]any{
			"message": "id must be of type uuid",
			"error":   err.Error(),
		})
		return
	}

	user, err := app.db.Delete(id)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{
			"error": err.Error(),
		})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	var body UserRequest
	id := chi.URLParam(r, "id")

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		jsonResponse(w, http.StatusUnprocessableEntity, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Validate user input
	if validator.NotBlank(body.FirstName) {
		body.Validator.CheckField(validator.MinChars(body.FirstName, 2) && validator.MaxChars(body.FirstName, 20), "first_name", "first_name must be at least 2 chars long and smaller than 20 chars")
	}

	if validator.NotBlank(body.LastName) {
		body.Validator.CheckField(validator.MinChars(body.LastName, 2) && validator.MaxChars(body.LastName, 20), "last_name", "last_name must be at least 2 chars long and smaller than 20 chars")
	}

	if validator.NotBlank(body.Bio) {
		body.Validator.CheckField(validator.MinChars(body.Bio, 20) && validator.MaxChars(body.LastName, 450), "bio", "bio must be at least 20 chars long and smaller than 450 chars")
	}

	if !body.Validator.Valid() {
		jsonResponse(w, http.StatusBadRequest, body.Validator.FieldErrors)
		return
	}

	user, err := app.db.Update(id, store.User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Bio:       body.Bio,
	})
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]any{
			"error": err.Error(),
		})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (app *application) bindUserRouter() *chi.Mux {
	userRouter := chi.NewRouter()

	userRouter.Post("/", app.addUser)
	userRouter.Get("/", app.getAllUsers)
	userRouter.Get("/{id}", app.getUserById)
	userRouter.Delete("/{id}", app.deleteUserById)
	userRouter.Put("/{id}", app.updateUser)

	return userRouter
}

func (app *application) bindRoutes() {
	apiRouter := chi.NewRouter()
	apiRouter.Mount("/users", app.bindUserRouter())

	app.router.Mount("/api", apiRouter)
}
