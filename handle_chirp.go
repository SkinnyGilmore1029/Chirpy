package main

import (
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
