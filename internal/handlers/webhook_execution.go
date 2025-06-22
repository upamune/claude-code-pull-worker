package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/models"
	"github.com/upamune/claude-code-pull-worker/internal/types"
)

type WebhookExecutionHandler struct {
	queries  *db.Queries
}

func NewWebhookExecutionHandler(queries *db.Queries) *WebhookExecutionHandler {
	return &WebhookExecutionHandler{
		queries:  queries,
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// truncateAPIKey safely truncates an API key for logging
func truncateAPIKey(apiKey string) string {
	if len(apiKey) <= 10 {
		return apiKey
	}
	return apiKey[:10] + "..."
}

func (h *WebhookExecutionHandler) HandleWebhookExecution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	webhookID := vars["uuid"]
	
	// Get webhook
	webhook, err := h.queries.GetWebhook(ctx, webhookID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Get all API keys for this webhook
	fullKeys, err := h.queries.GetAPIKeysForWebhook(ctx, webhookID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// If webhook has API keys configured, authentication is required
	var apiKeyID *int64
	if len(fullKeys) > 0 {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Log missing API key attempt
			go h.queries.LogSecurityAuditEvent(context.Background(), db.LogSecurityAuditEventParams{
				WebhookID:      webhookID,
				EventType:      "missing_api_key",
				ClientIp:       getClientIP(r),
				UserAgent:      sql.NullString{String: r.Header.Get("User-Agent"), Valid: true},
				ApiKeyProvided: sql.NullString{},
				ErrorMessage:   sql.NullString{String: "API key required but not provided", Valid: true},
				RequestPath:    sql.NullString{String: r.URL.Path, Valid: true},
			})
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}
		
		// Check each key by verifying the provided API key against stored hashes
		authorized := false
		for _, key := range fullKeys {
			// Compare the provided API key with the stored hash
			if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey)); err == nil {
				// Found a matching key
				authorized = true
				apiKeyID = &key.ID
				// Update last used timestamp
				go h.queries.UpdateAPIKeyLastUsed(context.Background(), key.ID)
				break
			}
		}
		
		if !authorized {
			// Log invalid API key attempt
			go h.queries.LogSecurityAuditEvent(context.Background(), db.LogSecurityAuditEventParams{
				WebhookID:      webhookID,
				EventType:      "invalid_api_key",
				ClientIp:       getClientIP(r),
				UserAgent:      sql.NullString{String: r.Header.Get("User-Agent"), Valid: true},
				ApiKeyProvided: sql.NullString{String: truncateAPIKey(apiKey), Valid: true},
				ErrorMessage:   sql.NullString{String: "Invalid API key provided", Valid: true},
				RequestPath:    sql.NullString{String: r.URL.Path, Valid: true},
			})
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
	}
	
	// Parse request
	var req models.WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}
	
	// Parse Claude options
	var claudeOpts types.ClaudeOptions
	if claudeBytes, ok := webhook.ClaudeOptions.([]byte); ok {
		if err := json.Unmarshal(claudeBytes, &claudeOpts); err != nil {
			// Use default options if parsing fails
			claudeOpts = types.ClaudeOptions{}
		}
	} else if claudeStr, ok := webhook.ClaudeOptions.(string); ok {
		if err := json.Unmarshal([]byte(claudeStr), &claudeOpts); err != nil {
			// Use default options if parsing fails
			claudeOpts = types.ClaudeOptions{}
		}
	}
	
	// Enqueue the job
	claudeOptsJSON, _ := json.Marshal(claudeOpts)
	job, err := h.queries.EnqueueJob(ctx, db.EnqueueJobParams{
		WebhookID:     webhookID,
		ApiKeyID:      func() sql.NullInt64 {
			if apiKeyID != nil {
				return sql.NullInt64{Int64: *apiKeyID, Valid: true}
			}
			return sql.NullInt64{}
		}(),
		Prompt:        req.Prompt,
		Context:       sql.NullString{String: req.Context, Valid: req.Context != ""},
		ClaudeOptions: sql.NullString{String: string(claudeOptsJSON), Valid: true},
		Priority:      0, // Default priority
	})
	
	if err != nil {
		http.Error(w, "Failed to enqueue job", http.StatusInternalServerError)
		return
	}
	
	// Return 200 with job information
	response := map[string]interface{}{
		"status":  "accepted",
		"message": "Webhook execution enqueued",
		"job_id":  job.ID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

