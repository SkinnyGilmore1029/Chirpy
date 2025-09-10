package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SkinnyGilmore1029/Chirpy/internal/auth"
	"github.com/SkinnyGilmore1029/Chirpy/internal/database"
	"github.com/google/uuid"
)

// define a struct to use
type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// def a struct for the response without the password?
// go
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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

	// make sure password is set and not default
	if req.Password == "unset" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// make the string into a hash
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	newUser, err := cfg.queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hash,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}
	// create the new user
	resp := userResponse{
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

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	gettoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	userId, err := auth.ValidateJWT(gettoken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	// userId is now available for use
	var req updateUserRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to retrieve request", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and Password are required", nil)
		return
	}

	hashpw, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// update user with new hashpw
	updateuser, err := cfg.queries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          req.Email,
		HashedPassword: hashpw,
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Cant Authorize", err)
		return
	}
	// respond with the updated user information
	resp := userResponse{
		ID:        updateuser.ID,
		CreatedAt: updateuser.CreatedAt,
		UpdatedAt: updateuser.UpdatedAt,
		Email:     updateuser.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}

}
