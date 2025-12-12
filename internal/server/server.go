package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	db "github.com/Daankoo/lab5_PT/db/sqlc"
)

type Server struct {
	router *chi.Mux
	store  db.Store
}

func NewServer(store db.Store) *Server {
	s := &Server{
		router: chi.NewRouter(),
		store:  store,
	}

	s.routes()

	return s
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)

	s.router.Route("/users", func(r chi.Router) {
		r.Post("/", s.handleCreateUser)
		r.Get("/", s.handleListUsers)
		r.Get("/{id}", s.handleGetUser)
		r.Put("/{id}", s.handleUpdateUser)
		r.Delete("/{id}", s.handleDeleteUser)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ==== USER HANDLERS ====

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int32  `json:"age"`
}

type validationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

// POST /users
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("JSON decode error:", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	errs := validateUserInput(req.Name, req.Email, req.Age)
	if len(errs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrorResponse{Errors: errs})
		return
	}

	arg := db.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	user, err := s.store.CreateUser(r.Context(), arg)
	if err != nil {
		log.Println("error creating user:", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GET /users
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		log.Println("error listing users:", err)
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GET /users/{id}
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Println("error getting user:", err)
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// PUT /users/{id}
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	errs := validateUserInput(req.Name, req.Email, req.Age)
	if len(errs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrorResponse{Errors: errs})
		return
	}

	arg := db.UpdateUserParams{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	user, err := s.store.UpdateUser(r.Context(), arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Println("error updating user:", err)
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DELETE /users/{id}
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.store.DeleteUser(r.Context(), id); err != nil {
		log.Println("error deleting user:", err)
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDParam(r *http.Request) (int32, error) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id64), nil
}

func (s *Server) Run(port string) {
	log.Println("Starting server on", port)
	log.Fatal(http.ListenAndServe(port, s.router))
}
