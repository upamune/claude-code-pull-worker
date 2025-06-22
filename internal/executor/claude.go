package executor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	claude "github.com/upamune/claude-code-go"
	"github.com/upamune/claude-code-pull-worker/internal/db"
)

type ClaudeExecutor struct {
	timeout time.Duration
}

func NewClaudeExecutor(timeout time.Duration) *ClaudeExecutor {
	return &ClaudeExecutor{
		timeout: timeout,
	}
}

// ExecuteWithOptions executes Claude with specific options from job
func (e *ClaudeExecutor) ExecuteWithOptions(ctx context.Context, prompt string, job db.JobQueue) (string, error) {
	opts := &claude.Options{
		WorkingDir:          job.WorkingDir.String,
		MaxThinkingTokens:   intPtrFromNullInt64(job.MaxThinkingTokens),
		MaxTurns:            intPtrFromNullInt64(job.MaxTurns),
		CustomSystemPrompt:  job.CustomSystemPrompt.String,
		AppendSystemPrompt:  job.AppendSystemPrompt.String,
		Model:               job.Model.String,
		FallbackModel:       job.FallbackModel.String,
	}

	// Parse comma-separated tool lists
	if job.AllowedTools.Valid && job.AllowedTools.String != "" {
		opts.AllowedTools = strings.Split(job.AllowedTools.String, ",")
		for i := range opts.AllowedTools {
			opts.AllowedTools[i] = strings.TrimSpace(opts.AllowedTools[i])
		}
	}

	if job.DisallowedTools.Valid && job.DisallowedTools.String != "" {
		opts.DisallowedTools = strings.Split(job.DisallowedTools.String, ",")
		for i := range opts.DisallowedTools {
			opts.DisallowedTools[i] = strings.TrimSpace(opts.DisallowedTools[i])
		}
	}

	// Set permission mode
	if job.PermissionMode.Valid && job.PermissionMode.String != "" {
		switch job.PermissionMode.String {
		case "allow":
			opts.PermissionMode = claude.PermissionBypassPermissions
		case "ask":
			opts.PermissionMode = claude.PermissionDefault
		case "deny":
			// deny is not directly supported, use default with caution
			opts.PermissionMode = claude.PermissionDefault
		// "auto" and "review" are not supported by the library, default to default
		case "auto", "review":
			opts.PermissionMode = claude.PermissionDefault
		}
	}

	if job.PermissionPromptToolName.Valid {
		opts.PermissionPromptToolName = job.PermissionPromptToolName.String
	}

	// TODO: Parse MCP servers from job.McpServers if needed

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Log execution details for debugging
	fmt.Printf("Executing claude with options: WorkingDir=%s, Model=%s, Prompt=%s\n", 
		opts.WorkingDir, opts.Model, prompt)
	
	// Execute
	result, err := claude.Query(ctx, prompt, opts)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timeout after %v", e.timeout)
		}
		// Log the full error details
		fmt.Printf("Claude execution error: %+v\n", err)
		return "", fmt.Errorf("execution error: %v", err)
	}

	return result.Result, nil
}

// Execute is a simple wrapper for backward compatibility
func (e *ClaudeExecutor) Execute(prompt string) (string, error) {
	ctx := context.Background()
	
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	result, err := claude.Query(ctx, prompt, nil)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timeout after %v", e.timeout)
		}
		return "", fmt.Errorf("execution error: %v", err)
	}

	return result.Result, nil
}

func intPtrFromNullInt64(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	val := int(n.Int64)
	return &val
}