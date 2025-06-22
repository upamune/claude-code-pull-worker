package handlers

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

type createAPIKeyRequest struct {
	Description string `json:"description"`
}

type apiKeyResponse struct {
	ID          int64  `json:"id"`
	APIKey      string `json:"api_key,omitempty"` // Only included on creation
	KeyPrefix   string `json:"key_prefix"`
	KeySuffix   string `json:"key_suffix"`
	Description string `json:"description"`
}

func generateAPIKey() (string, error) {
	// Generate 32 bytes of random data
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	
	// Encode to base64 and add prefix
	key := fmt.Sprintf("claude_%s", base64.URLEncoding.EncodeToString(randomBytes))
	// Remove padding
	key = strings.TrimRight(key, "=")
	
	return key, nil
}

func (h *AdminHandler) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	keys, err := h.queries.ListAPIKeysByWebhook(r.Context(), webhookID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the API key list
	var buf bytes.Buffer
	content, err := templates.GetFile(templates.APIKeyListItemTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.New("apikey").Parse(string(content)))
	
	for _, key := range keys {
		data := map[string]interface{}{
			"ID":          key.ID,
			"KeyPrefix":   key.KeyPrefix,
			"KeySuffix":   key.KeySuffix,
			"Description": key.Description.String,
			"LastUsedAt":  "Never",
		}
		
		if key.LastUsedAt.Valid {
			data["LastUsedAt"] = key.LastUsedAt.Time.Format("2006-01-02 15:04")
		}
		
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func (h *AdminHandler) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	var req createAPIKeyRequest
	
	// Handle form data
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Description = r.FormValue("description")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	// Hash the key
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash API key", http.StatusInternalServerError)
		return
	}

	// Extract prefix and suffix for display
	keyPrefix := apiKey[:10]
	keySuffix := apiKey[len(apiKey)-4:]

	// Create the key
	key, err := h.queries.CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		WebhookID:   webhookID,
		KeyHash:     string(hashedKey),
		KeyPrefix:   keyPrefix,
		KeySuffix:   keySuffix,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If it's an HTMX request, return the new key display
	if r.Header.Get("HX-Request") == "true" {
		// First show the full key
		var buf bytes.Buffer
		content, err := templates.GetFile(templates.NewAPIKeyResponseTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl := template.Must(template.New("newkey").Parse(string(content)))
		tmpl.Execute(&buf, map[string]string{"APIKey": apiKey})
		
		// Then append the updated list
		keys, err := h.queries.ListAPIKeysByWebhook(r.Context(), webhookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		content, err = templates.GetFile(templates.APIKeyListItemTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		keyTmpl := template.Must(template.New("apikey").Parse(string(content)))
		for _, k := range keys {
			data := map[string]interface{}{
				"ID":          k.ID,
				"KeyPrefix":   k.KeyPrefix,
				"KeySuffix":   k.KeySuffix,
				"Description": k.Description.String,
				"LastUsedAt":  "Never",
			}
			
			if k.LastUsedAt.Valid {
				data["LastUsedAt"] = k.LastUsedAt.Time.Format("2006-01-02 15:04")
			}
			
			if err := keyTmpl.Execute(&buf, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		
		w.Header().Set("Content-Type", "text/html")
		w.Write(buf.Bytes())
		return
	}

	// Otherwise return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKeyResponse{
		ID:          key.ID,
		APIKey:      apiKey,
		KeyPrefix:   key.KeyPrefix,
		KeySuffix:   key.KeySuffix,
		Description: key.Description.String,
	})
}

func (h *AdminHandler) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["id"]
	
	// Convert string to int64
	var id int64
	if _, err := fmt.Sscanf(keyID, "%d", &id); err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}
	
	err := h.queries.DeleteAPIKey(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If it's an HTMX request, return empty (element will be removed)
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}