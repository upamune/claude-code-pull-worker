package templates

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed html/*.html *.html
var templateFS embed.FS

// ParseFS parses all templates from the embedded filesystem
func ParseFS() (*template.Template, error) {
	return template.ParseFS(templateFS, "html/*.html", "*.html")
}

// GetFile reads a file from the embedded filesystem
func GetFile(name string) ([]byte, error) {
	return templateFS.ReadFile(name)
}

// GetFileFS returns a sub filesystem for a specific path
func GetFileFS(dir string) (fs.FS, error) {
	return fs.Sub(templateFS, dir)
}

// Template names for easy access
const (
	AdminTemplate              = "admin.html"
	WebhookDetailTemplate      = "webhook_detail.html"
	WebhookListItemTemplate    = "html/webhook_list_item.html"
	APIKeyListItemTemplate     = "html/api_key_list_item.html"
	GlobalSettingsFormTemplate = "html/global_settings_form.html"
	NewAPIKeyResponseTemplate  = "html/new_api_key_response.html"
	SecurityAuditLogItemTemplate = "html/security_audit_log_item.html"
	JobQueueItemTemplate       = "html/job_queue_item.html"
)