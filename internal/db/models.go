// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
	"time"
)

type ApiKey struct {
	ID          int64          `json:"id"`
	WebhookID   string         `json:"webhook_id"`
	KeyHash     string         `json:"key_hash"`
	KeyPrefix   string         `json:"key_prefix"`
	KeySuffix   string         `json:"key_suffix"`
	Description sql.NullString `json:"description"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	LastUsedAt  sql.NullTime   `json:"last_used_at"`
}

type ExecutionHistory struct {
	ID              int64          `json:"id"`
	WebhookID       string         `json:"webhook_id"`
	ApiKeyID        sql.NullInt64  `json:"api_key_id"`
	Prompt          string         `json:"prompt"`
	Response        sql.NullString `json:"response"`
	Error           sql.NullString `json:"error"`
	Success         bool           `json:"success"`
	ExecutionTimeMs sql.NullInt64  `json:"execution_time_ms"`
	CreatedAt       time.Time      `json:"created_at"`
}

type GlobalSetting struct {
	SettingKey   string      `json:"setting_key"`
	SettingValue interface{} `json:"setting_value"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type JobQueue struct {
	ID                       int64          `json:"id"`
	WebhookID                string         `json:"webhook_id"`
	ApiKeyID                 sql.NullInt64  `json:"api_key_id"`
	Prompt                   string         `json:"prompt"`
	JobStatus                string         `json:"job_status"`
	Priority                 int64          `json:"priority"`
	RetryCount               int64          `json:"retry_count"`
	MaxRetries               int64          `json:"max_retries"`
	WorkerID                 sql.NullString `json:"worker_id"`
	VisibilityTimeout        sql.NullTime   `json:"visibility_timeout"`
	ErrorMessage             sql.NullString `json:"error_message"`
	Response                 sql.NullString `json:"response"`
	ExecutionTimeMs          sql.NullInt64  `json:"execution_time_ms"`
	CreatedAt                time.Time      `json:"created_at"`
	StartedAt                sql.NullTime   `json:"started_at"`
	CompletedAt              sql.NullTime   `json:"completed_at"`
	WorkingDir               sql.NullString `json:"working_dir"`
	MaxThinkingTokens        sql.NullInt64  `json:"max_thinking_tokens"`
	MaxTurns                 sql.NullInt64  `json:"max_turns"`
	CustomSystemPrompt       sql.NullString `json:"custom_system_prompt"`
	AppendSystemPrompt       sql.NullString `json:"append_system_prompt"`
	AllowedTools             sql.NullString `json:"allowed_tools"`
	DisallowedTools          sql.NullString `json:"disallowed_tools"`
	PermissionMode           sql.NullString `json:"permission_mode"`
	PermissionPromptToolName sql.NullString `json:"permission_prompt_tool_name"`
	Model                    sql.NullString `json:"model"`
	FallbackModel            sql.NullString `json:"fallback_model"`
	McpServers               sql.NullString `json:"mcp_servers"`
	EnableContinue           bool           `json:"enable_continue"`
	ContinueMinutes          int64          `json:"continue_minutes"`
}

type SecurityAuditLog struct {
	ID             int64          `json:"id"`
	WebhookID      string         `json:"webhook_id"`
	EventType      string         `json:"event_type"`
	ClientIp       string         `json:"client_ip"`
	UserAgent      sql.NullString `json:"user_agent"`
	ApiKeyProvided sql.NullString `json:"api_key_provided"`
	ErrorMessage   sql.NullString `json:"error_message"`
	RequestPath    sql.NullString `json:"request_path"`
	CreatedAt      time.Time      `json:"created_at"`
}

type Webhook struct {
	ID                       string         `json:"id"`
	Name                     string         `json:"name"`
	Description              sql.NullString `json:"description"`
	IsActive                 bool           `json:"is_active"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	WorkingDir               sql.NullString `json:"working_dir"`
	MaxThinkingTokens        sql.NullInt64  `json:"max_thinking_tokens"`
	MaxTurns                 sql.NullInt64  `json:"max_turns"`
	CustomSystemPrompt       sql.NullString `json:"custom_system_prompt"`
	AppendSystemPrompt       sql.NullString `json:"append_system_prompt"`
	AllowedTools             sql.NullString `json:"allowed_tools"`
	DisallowedTools          sql.NullString `json:"disallowed_tools"`
	PermissionMode           sql.NullString `json:"permission_mode"`
	PermissionPromptToolName sql.NullString `json:"permission_prompt_tool_name"`
	Model                    sql.NullString `json:"model"`
	FallbackModel            sql.NullString `json:"fallback_model"`
	McpServers               sql.NullString `json:"mcp_servers"`
	NotificationConfig       interface{}    `json:"notification_config"`
	EnableContinue           bool           `json:"enable_continue"`
	ContinueMinutes          int64          `json:"continue_minutes"`
}
