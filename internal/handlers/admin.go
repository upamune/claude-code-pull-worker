package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

type AdminHandler struct {
	queries *db.Queries
}

func NewAdminHandler(queries *db.Queries) (*AdminHandler, error) {
	return &AdminHandler{
		queries: queries,
	}, nil
}

func (h *AdminHandler) RegisterRoutes(r *mux.Router) {
	// Admin UI routes
	r.HandleFunc("/", h.handleAdminIndex).Methods("GET")
	r.HandleFunc("/webhooks/{id}", h.handleWebhookDetail).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	
	// Webhook management
	api.HandleFunc("/webhooks", h.handleListWebhooks).Methods("GET")
	api.HandleFunc("/webhooks", h.handleCreateWebhook).Methods("POST")
	api.HandleFunc("/webhooks/{id}", h.handleGetWebhook).Methods("GET")
	api.HandleFunc("/webhooks/{id}", h.handleUpdateWebhook).Methods("PUT")
	api.HandleFunc("/webhooks/{id}", h.handleDeleteWebhook).Methods("DELETE")
	
	// API key management
	api.HandleFunc("/webhooks/{id}/keys", h.handleListAPIKeys).Methods("GET")
	api.HandleFunc("/webhooks/{id}/keys", h.handleCreateAPIKey).Methods("POST")
	api.HandleFunc("/keys/{id}", h.handleDeleteAPIKey).Methods("DELETE")
	
	// Execution history
	api.HandleFunc("/webhooks/{id}/executions", h.handleListExecutions).Methods("GET")
	api.HandleFunc("/webhooks/{id}/stats", h.handleGetStats).Methods("GET")
	
	// Job queue
	api.HandleFunc("/webhooks/{id}/queue", h.handleListJobQueue).Methods("GET")
	
	// Security logs
	api.HandleFunc("/webhooks/{id}/security-logs", h.handleListSecurityLogs).Methods("GET")
	
	// Global settings
	api.HandleFunc("/settings", h.handleGetSettings).Methods("GET")
	api.HandleFunc("/settings", h.handleUpdateSettings).Methods("PUT")
}

func (h *AdminHandler) handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	content, err := templates.GetFile(templates.AdminTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (h *AdminHandler) handleWebhookDetail(w http.ResponseWriter, r *http.Request) {
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

	// Render the webhook detail page
	content, err := templates.GetFile(templates.WebhookDetailTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmpl, err := template.New("webhook_detail").Parse(string(content))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ID":                       webhook.ID,
		"Name":                     webhook.Name,
		"Description":              webhook.Description.String,
		"IsActive":                 webhook.IsActive,
		"CreatedAt":                webhook.CreatedAt.Format("2006-01-02 15:04:05"),
		"WorkingDir":               webhook.WorkingDir.String,
		"MaxThinkingTokens":        func() string {
			if webhook.MaxThinkingTokens.Valid {
				return strconv.FormatInt(webhook.MaxThinkingTokens.Int64, 10)
			}
			return ""
		}(),
		"MaxTurns":                 func() string {
			if webhook.MaxTurns.Valid {
				return strconv.FormatInt(webhook.MaxTurns.Int64, 10)
			}
			return ""
		}(),
		"CustomSystemPrompt":       webhook.CustomSystemPrompt.String,
		"AppendSystemPrompt":       webhook.AppendSystemPrompt.String,
		"AllowedTools":             webhook.AllowedTools.String,
		"DisallowedTools":          webhook.DisallowedTools.String,
		"PermissionMode":           webhook.PermissionMode.String,
		"PermissionPromptToolName": webhook.PermissionPromptToolName.String,
		"Model":                    webhook.Model.String,
		"FallbackModel":            webhook.FallbackModel.String,
		"MCPServers":               webhook.McpServers.String,
		"NotificationConfig":       "",
		"DiscordWebhookURL":        "",
	}
	
	// Extract Discord webhook URL from notification config
	if notifBytes, ok := webhook.NotificationConfig.([]byte); ok {
		data["NotificationConfig"] = string(notifBytes)
		
		var notifConfig map[string]interface{}
		if err := json.Unmarshal(notifBytes, &notifConfig); err == nil {
			if discord, ok := notifConfig["discord"].(map[string]interface{}); ok {
				if webhookURL, ok := discord["webhook_url"].(string); ok {
					data["DiscordWebhookURL"] = webhookURL
				}
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}