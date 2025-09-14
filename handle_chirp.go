package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/SkinnyGilmore1029/Chirpy/internal/auth"
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

	// --- Authenticate user ---
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	// --- Decode request body ---
	var in newChirp
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// --- Validate chirp body length ---
	if len(in.Body) > 140 {
		http.Error(w, "chirp is too long", http.StatusBadRequest)
		return
	}

	// --- Clean bad words ---
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleanedBody := getCleanedBody(in.Body, badWords)

	// --- Create chirp in database ---
	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   cleanedBody,
		UserID: userID, // ✅ use user ID from JWT, not request body
	})
	if err != nil {
		http.Error(w, "could not create chirp", http.StatusInternalServerError)
		return
	}

	// --- Map DB model to response ---
	resp := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	// --- Send response ---
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
	// declare what chirps is first and error value to be changed later
	var chirps []database.Chirp
	var err error

	// check for author id in URL
	authId := r.URL.Query().Get("author_id")
	if authId == "" {
		// If no author ID is provided, return all chirps
		chirps, err = cfg.queries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps", err)
			return
		}
	} else {
		// If author ID is provided, return chirps for that author
		uid, err := uuid.Parse(authId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid authorId", err)
			return
		}

		chirps, err = cfg.queries.GetChirpsByAuthor(r.Context(), uid)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps", err)
			return
		}
	}

	// check for sort query parameter
	sortParam := r.URL.Query().Get("sort")
	if sortParam == "desc" {
		// sort descending by CreatedAt
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		// default is ascending
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}

	// make an array of chirpResponses to hold the chirps
	resp := make([]chirpResponse, 0, len(chirps))

	// loop through all chirps from the database
	for _, c := range chirps {
		resp = append(resp, chirpResponse{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserId:    c.UserID,
		})
	}

	// respond with the newly created JSON structs
	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	// Extract chirpID from the URL
	chirpId := r.PathValue("chirpID")

	// convert to uuid
	uid, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirpId", err)
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// get tokekn for the user
	userToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to Get token", err)
		return
	}

	// get user id from token
	userId, err := auth.ValidateJWT(userToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Extract chirpID from the URL
	chirpId := r.PathValue("chirpID")

	// convert to uuid
	uid, err := uuid.Parse(chirpId)
	if err != nil {
		http.Error(w, "invalid chirpID", http.StatusBadRequest)
		return
	}

	//chirp to be deleted
	chirp, err := cfg.queries.GetChirp(r.Context(), uid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	// Check if the chirp belongs to the user
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp", nil)
		return
	}

	// Delete the chirp
	if err := cfg.queries.RemoveChirp(r.Context(), uid); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
		return
	}

	// Respond with no content
	w.WriteHeader(http.StatusNoContent)
}
