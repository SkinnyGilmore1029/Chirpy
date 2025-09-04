package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// define a struct to use
type createUserRequest struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// create a user instance of struct User to decode into
	var req createUserRequest
	// decode the request body into the user instance
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	newUser, err := cfg.queries.CreateUser(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}
	// create the new user
	resp := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
