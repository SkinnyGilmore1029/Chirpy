package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := cfg.queries.ResetUsers(r.Context()); err != nil {
		http.Error(w, "Failed to reset users", http.StatusInternalServerError)
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
