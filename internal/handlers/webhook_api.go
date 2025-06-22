package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

type createWebhookRequest struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ClaudeOptions  json.RawMessage `json:"claude_options"`
	NotificationConfig json.RawMessage `json:"notification_config"`
}

func (h *AdminHandler) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	webhooks, err := h.queries.ListWebhooks(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the webhook list items
	var buf bytes.Buffer
	content, err := templates.GetFile(templates.WebhookListItemTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.New("webhook").Parse(string(content)))
	
	for _, webhook := range webhooks {
		// Get stats for each webhook
		stats, _ := h.queries.GetWebhookWithStats(ctx, webhook.ID)
		
		data := map[string]interface{}{
			"ID":             webhook.ID,
			"Name":           webhook.Name,
			"Description":    webhook.Description,
			"CreatedAt":      webhook.CreatedAt.Format("2006-01-02 15:04"),
			"ClaudeOptions":  func() string {
				if claudeBytes, ok := webhook.ClaudeOptions.([]byte); ok {
					return string(claudeBytes)
				}
				if claudeStr, ok := webhook.ClaudeOptions.(string); ok {
					return claudeStr
				}
				return "{}"
			}(),
			"APIKeyCount":    0,
			"ExecutionCount": 0,
			"LastExecution":  "Never",
		}
		
		if stats.ID != "" {
			data["APIKeyCount"] = stats.ApiKeyCount
			data["ExecutionCount"] = stats.ExecutionCount
			if lastExecTime, ok := stats.LastExecution.(time.Time); ok && !lastExecTime.IsZero() {
				data["LastExecution"] = lastExecTime.Format("2006-01-02 15:04")
			}
		}
		
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func (h *AdminHandler) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req createWebhookRequest
	
	// Handle form data
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		req.Name = r.FormValue("name")
		req.Description = r.FormValue("description")
		
		// Parse JSON options
		optionsStr := r.FormValue("claude_options")
		if optionsStr == "" {
			req.ClaudeOptions = json.RawMessage("{}")
		} else {
			req.ClaudeOptions = json.RawMessage(optionsStr)
		}
		
		notifStr := r.FormValue("notification_config")
		if notifStr != "" {
			req.NotificationConfig = json.RawMessage(notifStr)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Generate UUID
	id := uuid.New().String()

	// Create webhook
	webhook, err := h.queries.CreateWebhook(r.Context(), db.CreateWebhookParams{
		ID:            id,
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		ClaudeOptions: req.ClaudeOptions,
		NotificationConfig: req.NotificationConfig,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If it's an HTMX request, return the updated list
	if r.Header.Get("HX-Request") == "true" {
		h.handleListWebhooks(w, r)
		return
	}

	// Otherwise return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

func (h *AdminHandler) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhook, err := h.queries.GetWebhook(r.Context(), vars["id"])
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

func (h *AdminHandler) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req createWebhookRequest
	
	// Handle form data
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		req.Name = r.FormValue("name")
		req.Description = r.FormValue("description")
		req.ClaudeOptions = json.RawMessage(r.FormValue("claude_options"))
		req.NotificationConfig = json.RawMessage(r.FormValue("notification_config"))
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	err := h.queries.UpdateWebhook(r.Context(), db.UpdateWebhookParams{
		Name:           req.Name,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		ClaudeOptions:  req.ClaudeOptions,
		NotificationConfig: req.NotificationConfig,
		ID:             vars["id"],
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := h.queries.DeleteWebhook(r.Context(), vars["id"])
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