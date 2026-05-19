package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type PushTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (r *router) registerPushToken(w http.ResponseWriter, req *http.Request, authToken string) {
	var input PushTokenRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	token := strings.TrimSpace(input.Token)
	platform := strings.ToLower(strings.TrimSpace(input.Platform))
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}
	if platform == "" {
		platform = "ios"
	}
	if platform != "ios" && platform != "android" {
		writeError(w, http.StatusBadRequest, "platform must be ios or android")
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	q := make(map[string][]string)
	q["on_conflict"] = []string{"token"}
	payload := map[string]any{
		"token":        token,
		"user_id":      req.Header.Get("X-Nomo-User-ID"),
		"platform":     platform,
		"updated_at":   now,
		"last_seen_at": now,
	}
	var rows []map[string]any
	if err := r.deps.Supabase.Upsert(req.Context(), authToken, "push_tokens", q, payload, &rows); err != nil {
		writeSupabaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, firstMap(rows, payload))
}
