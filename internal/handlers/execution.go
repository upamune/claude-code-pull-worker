package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/upamune/claude-code-pull-worker/internal/db"
)

type executionListItem struct {
	ID            int64     `json:"id"`
	Prompt        string    `json:"prompt"`
	Success       bool      `json:"success"`
	ExecutionTime int       `json:"execution_time_ms"`
	CreatedAt     time.Time `json:"created_at"`
	Error         string    `json:"error,omitempty"`
}

func (h *AdminHandler) handleListExecutions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	// Parse pagination
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	offset := (page - 1) * limit
	
	// Get executions
	executions, err := h.queries.ListExecutionHistoriesByWebhook(r.Context(), db.ListExecutionHistoriesByWebhookParams{
		WebhookID: webhookID,
		Limit:     int64(limit),
		Offset:    int64(offset),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get total count
	count, err := h.queries.CountExecutionHistoriesByWebhook(r.Context(), webhookID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Render HTML for HTMX
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		
		if len(executions) == 0 {
			w.Write([]byte(`<div class="text-gray-500 text-center py-8">No executions yet</div>`))
			return
		}
		
		// Render execution list
		fmt.Fprintf(w, `<div class="space-y-4">`)
		for _, exec := range executions {
			statusClass := "bg-green-100 text-green-800"
			statusText := "Success"
			if !exec.Success {
				statusClass = "bg-red-100 text-red-800"
				statusText = "Failed"
			}
			
			fmt.Fprintf(w, `
			<div class="bg-white rounded-lg shadow p-4">
				<div class="flex justify-between items-start">
					<div class="flex-1">
						<div class="flex items-center gap-2 mb-2">
							<span class="px-2 py-1 rounded-full text-xs font-medium %s">%s</span>
							<span class="text-sm text-gray-500">%s</span>
							<span class="text-sm text-gray-500">%dms</span>
						</div>
						<div class="text-sm font-mono bg-gray-100 p-2 rounded">%s</div>
						%s
					</div>
				</div>
			</div>
			`, statusClass, statusText, 
			exec.CreatedAt.Format("2006-01-02 15:04:05"),
			exec.ExecutionTimeMs.Int64,
			escapeHTML(exec.Prompt),
			func() string {
				if exec.Error.Valid && exec.Error.String != "" {
					return fmt.Sprintf(`<div class="mt-2 text-sm text-red-600">%s</div>`, escapeHTML(exec.Error.String))
				}
				return ""
			}())
		}
		fmt.Fprintf(w, `</div>`)
		
		// Pagination
		if count > int64(limit) {
			totalPages := int((count + int64(limit) - 1) / int64(limit))
			fmt.Fprintf(w, `<div class="mt-6 flex justify-center gap-2">`)
			
			for i := 1; i <= totalPages; i++ {
				activeClass := ""
				if i == page {
					activeClass = "bg-blue-600 text-white"
				} else {
					activeClass = "bg-white text-gray-700 hover:bg-gray-50"
				}
				
				fmt.Fprintf(w, `
				<button hx-get="/api/webhooks/%s/executions?page=%d" 
					hx-target="[hx-get='/api/webhooks/%s/executions']"
					hx-swap="innerHTML"
					class="px-3 py-1 rounded %s">
					%d
				</button>`, webhookID, i, webhookID, activeClass, i)
			}
			
			fmt.Fprintf(w, `</div>`)
		}
		
		return
	}
	
	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"executions": executions,
		"total":      count,
		"page":       page,
		"limit":      limit,
	})
}

func (h *AdminHandler) handleGetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	webhookID := vars["id"]
	
	// Get stats for last 30 days
	since := time.Now().AddDate(0, 0, -30)
	
	stats, err := h.queries.GetExecutionStats(r.Context(), db.GetExecutionStatsParams{
		WebhookID: webhookID,
		CreatedAt: since,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Render HTML for HTMX
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		
		successRate := float64(0)
		if stats.TotalExecutions > 0 && stats.SuccessfulExecutions.Valid {
			successRate = stats.SuccessfulExecutions.Float64 / float64(stats.TotalExecutions) * 100
		}
		
		avgTime := float64(0)
		if stats.AvgExecutionTimeMs.Valid {
			avgTime = stats.AvgExecutionTimeMs.Float64
		}
		
		fmt.Fprintf(w, `
		<div class="bg-white rounded-lg shadow p-6">
			<h3 class="text-lg font-semibold mb-4">Last 30 Days Statistics</h3>
			<div class="grid grid-cols-3 gap-6">
				<div>
					<div class="text-2xl font-bold">%d</div>
					<div class="text-sm text-gray-600">Total Executions</div>
				</div>
				<div>
					<div class="text-2xl font-bold">%.1f%%</div>
					<div class="text-sm text-gray-600">Success Rate</div>
				</div>
				<div>
					<div class="text-2xl font-bold">%.0fms</div>
					<div class="text-sm text-gray-600">Avg Execution Time</div>
				</div>
			</div>
		</div>
		`, stats.TotalExecutions, successRate, avgTime)
		
		return
	}
	
	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func escapeHTML(s string) string {
	// Simple HTML escaping
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}