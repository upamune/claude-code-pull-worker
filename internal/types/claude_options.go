package types

// PermissionMode defines how Claude Code handles permissions
type PermissionMode string

const (
	PermissionModeAllow PermissionMode = "allow"
	PermissionModeAsk   PermissionMode = "ask"
)

// MCPStdioServerConfig configures an MCP stdio server
type MCPStdioServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// MCPSSEServerConfig configures an MCP SSE server
type MCPSSEServerConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// MCPHTTPServerConfig configures an MCP HTTP server
type MCPHTTPServerConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// MCPServerConfig can be one of the MCP server types
type MCPServerConfig interface{}

// ClaudeOptions configures the behavior of Claude Code SDK
type ClaudeOptions struct {
	// Tools configuration
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`

	// System prompt configuration
	CustomSystemPrompt string `json:"custom_system_prompt,omitempty"`
	AppendSystemPrompt string `json:"append_system_prompt,omitempty"`

	// Working directory for the Claude Code CLI
	WorkingDir string `json:"working_dir,omitempty"`

	// Token and turn limits
	MaxThinkingTokens *int `json:"max_thinking_tokens,omitempty"`
	MaxTurns          *int `json:"max_turns,omitempty"`

	// MCP server configuration
	MCPServers map[string]MCPServerConfig `json:"mcp_servers,omitempty"`

	// Path to the Claude Code CLI executable
	PathToClaudeCodeExecutable string `json:"path_to_claude_code_executable,omitempty"`

	// Permission handling
	PermissionMode           PermissionMode `json:"permission_mode,omitempty"`
	PermissionPromptToolName string         `json:"permission_prompt_tool_name,omitempty"`

	// Session continuation
	Continue bool   `json:"continue,omitempty"`
	Resume   string `json:"resume,omitempty"`

	// Model configuration
	Model         string `json:"model,omitempty"`
	FallbackModel string `json:"fallback_model,omitempty"`
}