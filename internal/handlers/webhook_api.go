package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

type createWebhookRequest struct {
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	NotificationConfig       json.RawMessage `json:"notification_config"`
	WorkingDir               string          `json:"working_dir"`
	MaxThinkingTokens        *int            `json:"max_thinking_tokens"`
	MaxTurns                 *int            `json:"max_turns"`
	CustomSystemPrompt       string          `json:"custom_system_prompt"`
	AppendSystemPrompt       string          `json:"append_system_prompt"`
	AllowedTools             string          `json:"allowed_tools"`
	DisallowedTools          string          `json:"disallowed_tools"`
	PermissionMode           string          `json:"permission_mode"`
	PermissionPromptToolName string          `json:"permission_prompt_tool_name"`
	Model                    string          `json:"model"`
	FallbackModel            string          `json:"fallback_model"`
	MCPServers               string          `json:"mcp_servers"`
	EnableContinue           bool            `json:"enable_continue"`
	ContinueMinutes          int             `json:"continue_minutes"`
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
		req.WorkingDir = r.FormValue("working_dir")
		req.CustomSystemPrompt = r.FormValue("custom_system_prompt")
		req.AppendSystemPrompt = r.FormValue("append_system_prompt")
		req.AllowedTools = r.FormValue("allowed_tools")
		req.DisallowedTools = r.FormValue("disallowed_tools")
		req.PermissionMode = r.FormValue("permission_mode")
		req.PermissionPromptToolName = r.FormValue("permission_prompt_tool_name")
		req.Model = r.FormValue("model")
		req.FallbackModel = r.FormValue("fallback_model")
		req.MCPServers = r.FormValue("mcp_servers")
		
		// Parse boolean enable_continue field
		req.EnableContinue = r.FormValue("enable_continue") == "true"
		
		// Parse integer fields
		if val := r.FormValue("max_thinking_tokens"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.MaxThinkingTokens = &n
			}
		}
		if val := r.FormValue("max_turns"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.MaxTurns = &n
			}
		}
		if val := r.FormValue("continue_minutes"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.ContinueMinutes = n
			}
		} else {
			req.ContinueMinutes = 10 // デフォルト値
		}
		
		// Handle Discord webhook URL
		discordWebhookURL := r.FormValue("discord_webhook_url")
		if discordWebhookURL != "" {
			notifConfig := map[string]interface{}{
				"discord": map[string]string{
					"webhook_url": discordWebhookURL,
				},
			}
			jsonBytes, _ := json.Marshal(notifConfig)
			req.NotificationConfig = json.RawMessage(jsonBytes)
		} else {
			// Use empty JSON if no webhook URL provided
			req.NotificationConfig = json.RawMessage("{}")
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
		ID:                       id,
		Name:                     req.Name,
		Description:              sql.NullString{String: req.Description, Valid: req.Description != ""},
		NotificationConfig:       req.NotificationConfig,
		WorkingDir:               sql.NullString{String: req.WorkingDir, Valid: req.WorkingDir != ""},
		MaxThinkingTokens:        func() sql.NullInt64 {
			if req.MaxThinkingTokens != nil {
				return sql.NullInt64{Int64: int64(*req.MaxThinkingTokens), Valid: true}
			}
			return sql.NullInt64{}
		}(),
		MaxTurns:                 func() sql.NullInt64 {
			if req.MaxTurns != nil {
				return sql.NullInt64{Int64: int64(*req.MaxTurns), Valid: true}
			}
			return sql.NullInt64{}
		}(),
		CustomSystemPrompt:       sql.NullString{String: req.CustomSystemPrompt, Valid: req.CustomSystemPrompt != ""},
		AppendSystemPrompt:       sql.NullString{String: req.AppendSystemPrompt, Valid: req.AppendSystemPrompt != ""},
		AllowedTools:             sql.NullString{String: req.AllowedTools, Valid: req.AllowedTools != ""},
		DisallowedTools:          sql.NullString{String: req.DisallowedTools, Valid: req.DisallowedTools != ""},
		PermissionMode:           sql.NullString{String: req.PermissionMode, Valid: req.PermissionMode != ""},
		PermissionPromptToolName: sql.NullString{String: req.PermissionPromptToolName, Valid: req.PermissionPromptToolName != ""},
		Model:                    sql.NullString{String: req.Model, Valid: req.Model != ""},
		FallbackModel:            sql.NullString{String: req.FallbackModel, Valid: req.FallbackModel != ""},
		McpServers:               sql.NullString{String: req.MCPServers, Valid: req.MCPServers != ""},
		EnableContinue:           req.EnableContinue,
		ContinueMinutes:          int64(req.ContinueMinutes),
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
		req.WorkingDir = r.FormValue("working_dir")
		req.CustomSystemPrompt = r.FormValue("custom_system_prompt")
		req.AppendSystemPrompt = r.FormValue("append_system_prompt")
		req.AllowedTools = r.FormValue("allowed_tools")
		req.DisallowedTools = r.FormValue("disallowed_tools")
		req.PermissionMode = r.FormValue("permission_mode")
		req.PermissionPromptToolName = r.FormValue("permission_prompt_tool_name")
		req.Model = r.FormValue("model")
		req.FallbackModel = r.FormValue("fallback_model")
		req.MCPServers = r.FormValue("mcp_servers")
		
		// Parse boolean enable_continue field
		req.EnableContinue = r.FormValue("enable_continue") == "true"
		
		// Parse integer fields
		if val := r.FormValue("max_thinking_tokens"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.MaxThinkingTokens = &n
			}
		}
		if val := r.FormValue("max_turns"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.MaxTurns = &n
			}
		}
		if val := r.FormValue("continue_minutes"); val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				req.ContinueMinutes = n
			}
		} else {
			req.ContinueMinutes = 10 // デフォルト値
		}
		
		// Handle Discord webhook URL
		discordWebhookURL := r.FormValue("discord_webhook_url")
		if discordWebhookURL != "" {
			notifConfig := map[string]interface{}{
				"discord": map[string]string{
					"webhook_url": discordWebhookURL,
				},
			}
			jsonBytes, _ := json.Marshal(notifConfig)
			req.NotificationConfig = json.RawMessage(jsonBytes)
		} else {
			// Use empty JSON if no webhook URL provided
			req.NotificationConfig = json.RawMessage("{}")
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	err := h.queries.UpdateWebhook(r.Context(), db.UpdateWebhookParams{
		Name:                     req.Name,
		Description:              sql.NullString{String: req.Description, Valid: req.Description != ""},
		NotificationConfig:       req.NotificationConfig,
		WorkingDir:               sql.NullString{String: req.WorkingDir, Valid: req.WorkingDir != ""},
		MaxThinkingTokens:        func() sql.NullInt64 {
			if req.MaxThinkingTokens != nil {
				return sql.NullInt64{Int64: int64(*req.MaxThinkingTokens), Valid: true}
			}
			return sql.NullInt64{}
		}(),
		MaxTurns:                 func() sql.NullInt64 {
			if req.MaxTurns != nil {
				return sql.NullInt64{Int64: int64(*req.MaxTurns), Valid: true}
			}
			return sql.NullInt64{}
		}(),
		CustomSystemPrompt:       sql.NullString{String: req.CustomSystemPrompt, Valid: req.CustomSystemPrompt != ""},
		AppendSystemPrompt:       sql.NullString{String: req.AppendSystemPrompt, Valid: req.AppendSystemPrompt != ""},
		AllowedTools:             sql.NullString{String: req.AllowedTools, Valid: req.AllowedTools != ""},
		DisallowedTools:          sql.NullString{String: req.DisallowedTools, Valid: req.DisallowedTools != ""},
		PermissionMode:           sql.NullString{String: req.PermissionMode, Valid: req.PermissionMode != ""},
		PermissionPromptToolName: sql.NullString{String: req.PermissionPromptToolName, Valid: req.PermissionPromptToolName != ""},
		Model:                    sql.NullString{String: req.Model, Valid: req.Model != ""},
		FallbackModel:            sql.NullString{String: req.FallbackModel, Valid: req.FallbackModel != ""},
		McpServers:               sql.NullString{String: req.MCPServers, Valid: req.MCPServers != ""},
		EnableContinue:           req.EnableContinue,
		ContinueMinutes:          int64(req.ContinueMinutes),
		ID:                       vars["id"],
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