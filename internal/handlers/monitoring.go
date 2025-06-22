package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/templates"
)

// handleListSecurityLogs returns security audit logs for a webhook
func (h *AdminHandler) handleListSecurityLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	// Get security logs (limit to recent 100)
	logs, err := h.queries.GetSecurityAuditLogs(r.Context(), db.GetSecurityAuditLogsParams{
		WebhookID: webhookID,
		Limit:     100,
		Offset:    0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Render the response
	w.Header().Set("Content-Type", "text/html")
	
	// Header
	w.Write([]byte(`
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200">
				<h3 class="text-lg font-medium text-gray-900">Security Audit Logs</h3>
				<p class="mt-1 text-sm text-gray-500">Unauthorized access attempts and API key validation failures</p>
			</div>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Event</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Client IP</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User Agent</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">API Key</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Error</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
	`))
	
	// Render each log item
	content, err := templates.GetFile(templates.SecurityAuditLogItemTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.New("security").Parse(string(content)))
	
	for _, log := range logs {
		data := map[string]interface{}{
			"CreatedAt":      log.CreatedAt.Format("2006-01-02 15:04:05"),
			"EventType":      log.EventType,
			"ClientIP":       log.ClientIp,
			"UserAgent":      log.UserAgent.String,
			"APIKeyProvided": log.ApiKeyProvided.String,
			"ErrorMessage":   log.ErrorMessage.String,
		}
		
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf.Bytes())
	}
	
	// Footer
	w.Write([]byte(`
					</tbody>
				</table>
			</div>
		</div>
	`))
}

// handleListJobQueue returns pending and processing jobs for a webhook
func (h *AdminHandler) handleListJobQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	// Get pending and processing jobs
	jobs, err := h.queries.GetJobsByWebhook(r.Context(), db.GetJobsByWebhookParams{
		WebhookID: webhookID,
		Limit:     50,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Render the response
	w.Header().Set("Content-Type", "text/html")
	
	// Header
	w.Write([]byte(`
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200">
				<h3 class="text-lg font-medium text-gray-900">Job Queue</h3>
				<p class="mt-1 text-sm text-gray-500">Pending and processing jobs for this webhook</p>
			</div>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Prompt</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Priority</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Started</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Completed</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
	`))
	
	// Render each job item
	content, err := templates.GetFile(templates.JobQueueItemTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.New("queue").Parse(string(content)))
	
	for _, job := range jobs {
		data := map[string]interface{}{
			"ID":          job.ID,
			"Status":      job.Status,
			"Prompt":      job.Prompt,
			"Priority":    job.Priority,
			"CreatedAt":   job.CreatedAt.Format("15:04:05"),
			"StartedAt":   job.StartedAt,
			"CompletedAt": job.CompletedAt,
		}
		
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf.Bytes())
	}
	
	// Footer with last update time
	w.Write([]byte(`
					</tbody>
				</table>
			</div>
			<div class="px-6 py-3 bg-gray-50 text-sm text-gray-500">
				Last updated: ` + time.Now().Format("15:04:05") + `
			</div>
		</div>
	`))
}