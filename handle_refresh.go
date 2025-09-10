package main

import (
	"net/http"
	"time"

	"github.com/SkinnyGilmore1029/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	gettoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	getuser, err := cfg.queries.GetUserFromRefreshToken(r.Context(), gettoken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	token, err := auth.MakeJWT(getuser.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create jwt", err)
		return
	}

	// If we reach this point, the token is valid and not expired
	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
