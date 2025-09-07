package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SkinnyGilmore1029/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// create a login Request
	var logreq loginRequest

	// decode the http.Request into logreq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logreq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to retrieve login request", err)
		return
	}
	// make sure email is valid
	if strings.TrimSpace(logreq.Email) == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	// make sure password is set and not default
	if logreq.Password == "unset" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}
	getUser, err := cfg.queries.GetUserByEmail(r.Context(), logreq.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if err := auth.CheckPasswordHash(logreq.Password, getUser.HashedPassword); err != nil {
		http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
		return
	}
	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:        getUser.ID,
		CreatedAt: getUser.CreatedAt,
		UpdatedAt: getUser.UpdatedAt,
		Email:     getUser.Email,
	})
}
