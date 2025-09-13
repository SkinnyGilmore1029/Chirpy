package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type PolkaWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserId uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handler_webhook(w http.ResponseWriter, r *http.Request) {
	//Parse Incoming Json Data
	var webhook PolkaWebhook
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&webhook); err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad request", err)
		return
	}
	if webhook.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	_, err := cfg.queries.UpgradeUser(r.Context(), webhook.Data.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
