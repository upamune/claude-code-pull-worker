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
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		apiKey := ""
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
		
		if apiKey == "" {
			// Log missing API key attempt
			go h.queries.LogSecurityAuditEvent(context.Background(), db.LogSecurityAuditEventParams{
				WebhookID:      webhookID,
				EventType:      "missing_api_key",
				ClientIp:       getClientIP(r),
				UserAgent:      sql.NullString{String: r.Header.Get("User-Agent"), Valid: true},
				ApiKeyProvided: sql.NullString{},
				ErrorMessage:   sql.NullString{String: "Authorization header with Bearer token required but not provided", Valid: true},
				RequestPath:    sql.NullString{String: r.URL.Path, Valid: true},
			})
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Authorization required", http.StatusUnauthorized)
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
	
	// Enqueue the job with Claude options from webhook
	job, err := h.queries.EnqueueJob(ctx, db.EnqueueJobParams{
		WebhookID:     webhookID,
		ApiKeyID:      func() sql.NullInt64 {
			if apiKeyID != nil {
				return sql.NullInt64{Int64: *apiKeyID, Valid: true}
			}
			return sql.NullInt64{}
		}(),
		Prompt:                   req.Prompt,
		Priority:                 0, // Default priority
		WorkingDir:               webhook.WorkingDir,
		MaxThinkingTokens:        webhook.MaxThinkingTokens,
		MaxTurns:                 webhook.MaxTurns,
		CustomSystemPrompt:       webhook.CustomSystemPrompt,
		AppendSystemPrompt:       webhook.AppendSystemPrompt,
		AllowedTools:             webhook.AllowedTools,
		DisallowedTools:          webhook.DisallowedTools,
		PermissionMode:           webhook.PermissionMode,
		PermissionPromptToolName: webhook.PermissionPromptToolName,
		Model:                    webhook.Model,
		FallbackModel:            webhook.FallbackModel,
		McpServers:               webhook.McpServers,
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

