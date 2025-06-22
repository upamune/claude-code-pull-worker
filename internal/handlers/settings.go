package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

type globalSettings struct {
	DiscordWebhookURL    string `json:"discord_webhook_url"`
	DefaultClaudeOptions string `json:"default_claude_options"`
}

func (h *AdminHandler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get notification config
	notifValue, err := h.queries.GetGlobalSetting(ctx, "default_notification_config")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get default claude options
	claudeValue, err := h.queries.GetGlobalSetting(ctx, "default_claude_options")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Parse notification config to get Discord URL
	var notifConfig map[string]interface{}
	notifBytes, ok := notifValue.([]byte)
	if !ok {
		// Handle string type
		if notifStr, ok := notifValue.(string); ok {
			notifBytes = []byte(notifStr)
		} else {
			http.Error(w, "Invalid notification config type", http.StatusInternalServerError)
			return
		}
	}
	if err := json.Unmarshal(notifBytes, &notifConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	discordURL := ""
	if discord, ok := notifConfig["discord"].(map[string]interface{}); ok {
		if url, ok := discord["webhook_url"].(string); ok {
			discordURL = url
		}
	}
	
	// Render HTML for HTMX
	if r.Header.Get("HX-Request") == "true" {
		var buf bytes.Buffer
		content, err := templates.GetFile(templates.GlobalSettingsFormTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl := template.Must(template.New("settings").Parse(string(content)))
		
		data := map[string]interface{}{
			"DiscordWebhookURL":    discordURL,
			"DefaultClaudeOptions": func() string {
				if claudeBytes, ok := claudeValue.([]byte); ok {
					return string(claudeBytes)
				}
				if claudeStr, ok := claudeValue.(string); ok {
					return claudeStr
				}
				return "{}"
			}(),
		}
		
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		w.Write(buf.Bytes())
		return
	}
	
	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globalSettings{
		DiscordWebhookURL:    discordURL,
		DefaultClaudeOptions: func() string {
			if claudeBytes, ok := claudeValue.([]byte); ok {
				return string(claudeBytes)
			}
			if claudeStr, ok := claudeValue.(string); ok {
				return claudeStr
			}
			return "{}"
		}(),
	})
}

func (h *AdminHandler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings globalSettings
	
	// Handle form data
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		settings.DiscordWebhookURL = r.FormValue("discord_webhook_url")
		settings.DefaultClaudeOptions = r.FormValue("default_claude_options")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	
	// Update notification config
	notifConfig := map[string]interface{}{
		"discord": map[string]string{
			"webhook_url": settings.DiscordWebhookURL,
		},
	}
	
	notifJSON, err := json.Marshal(notifConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if err := h.queries.UpdateGlobalSetting(r.Context(), db.UpdateGlobalSettingParams{
		Value: notifJSON,
		Key:   "default_notification_config",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update Claude options
	if err := h.queries.UpdateGlobalSetting(r.Context(), db.UpdateGlobalSettingParams{
		Value: json.RawMessage(settings.DefaultClaudeOptions),
		Key:   "default_claude_options",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}