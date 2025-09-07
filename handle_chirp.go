package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SkinnyGilmore1029/Chirpy/internal/database"
	"github.com/google/uuid"
)

type newChirp struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}
type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decode request body into struct
	var in newChirp
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate body length
	if len(in.Body) > 140 {
		http.Error(w, "chirp is too long", http.StatusBadRequest)
		return
	}

	// Clean bad words
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleanedBody := getCleanedBody(in.Body, badWords)

	// Parse user_id into UUID
	uid, err := uuid.Parse(in.UserId)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	// Create chirp in database
	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   cleanedBody,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "could not create chirp", http.StatusInternalServerError)
		return
	}

	// Map to response struct with snake_case
	resp := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	// Send response
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// Helper to replace bad words with ****
func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, w := range words {
		if _, exists := badWords[strings.ToLower(w)]; exists {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

// handler function to get all chirps
func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {

	//use the new generate function to get all the chirps from the database
	chirps, err := cfg.queries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps", err)
		return
	}
	// make an array of chirpsresponses to hold the chirps
	resp := make([]chirpResponse, 0, len(chirps))

	// Loop through all chirps from the database
	for _, c := range chirps {
		// append the chirpResponse into an array
		resp = append(resp, chirpResponse{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserId:    c.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	// Extract chirpID from the URL
	chirpId := r.PathValue("chirpID")

	// convert to uuid
	uid, err := uuid.Parse(chirpId)
	if err != nil {
		http.Error(w, "invalid chirpID", http.StatusBadRequest)
		return
	}

	// find the chirp in the database
	chirp, err := cfg.queries.GetChirp(r.Context(), uid)
	// make sure there isnt an error
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp", err)
		return

	}
	// make a response so the chirp has something to be loaded into
	resp := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
