-- Claude Code Pull Worker Database Schema
-- Complete schema with all tables and indexes

-- Create webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    working_dir TEXT,
    max_thinking_tokens INTEGER,
    max_turns INTEGER,
    custom_system_prompt TEXT,
    append_system_prompt TEXT,
    allowed_tools TEXT,
    disallowed_tools TEXT,
    permission_mode TEXT DEFAULT 'auto',
    permission_prompt_tool_name TEXT,
    model TEXT,
    fallback_model TEXT,
    mcp_servers TEXT,
    notification_config JSON,
    enable_continue BOOLEAN NOT NULL DEFAULT 1,
    continue_minutes INTEGER NOT NULL DEFAULT 10
);

-- Create api_keys table  
CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_suffix TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE
);

-- Create execution_histories table
CREATE TABLE IF NOT EXISTS execution_histories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL,
    api_key_id INTEGER,
    prompt TEXT NOT NULL,
    response TEXT,
    error TEXT,
    success BOOLEAN NOT NULL,
    execution_time_ms INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL
);

-- Create job_queue table
CREATE TABLE IF NOT EXISTS job_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL,
    api_key_id INTEGER,
    prompt TEXT NOT NULL,
    job_status TEXT DEFAULT 'pending' NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    worker_id TEXT,
    visibility_timeout DATETIME,
    error_message TEXT,
    response TEXT,
    execution_time_ms INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    working_dir TEXT,
    max_thinking_tokens INTEGER,
    max_turns INTEGER,
    custom_system_prompt TEXT,
    append_system_prompt TEXT,
    allowed_tools TEXT,
    disallowed_tools TEXT,
    permission_mode TEXT,
    permission_prompt_tool_name TEXT,
    model TEXT,
    fallback_model TEXT,
    mcp_servers TEXT,
    enable_continue BOOLEAN NOT NULL DEFAULT 0,
    continue_minutes INTEGER NOT NULL DEFAULT 10,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL
);

-- Create security_audit_logs table
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    user_agent TEXT,
    api_key_provided TEXT,
    error_message TEXT,
    request_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE
);

-- Create global_settings table
CREATE TABLE IF NOT EXISTS global_settings (
    setting_key TEXT PRIMARY KEY,
    setting_value JSON NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_api_keys_webhook_id ON api_keys(webhook_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_execution_histories_webhook_id ON execution_histories(webhook_id);
CREATE INDEX idx_execution_histories_created_at ON execution_histories(created_at);
CREATE INDEX idx_job_queue_status ON job_queue(job_status);
CREATE INDEX idx_job_queue_webhook_id ON job_queue(webhook_id);
CREATE INDEX idx_job_queue_created_at ON job_queue(created_at);
CREATE INDEX idx_job_queue_visibility_timeout ON job_queue(visibility_timeout);
CREATE INDEX idx_security_audit_logs_webhook_id ON security_audit_logs(webhook_id);
CREATE INDEX idx_security_audit_logs_created_at ON security_audit_logs(created_at);
CREATE INDEX idx_security_audit_logs_event_type ON security_audit_logs(event_type);
CREATE INDEX idx_security_audit_logs_client_ip ON security_audit_logs(client_ip);

