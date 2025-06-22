-- Create security_audit_logs table for tracking unauthorized access attempts
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL,
    event_type TEXT NOT NULL, -- 'missing_api_key', 'invalid_api_key', 'inactive_api_key'
    client_ip TEXT NOT NULL,
    user_agent TEXT,
    api_key_provided TEXT, -- Store partial key for debugging (first 10 chars only)
    error_message TEXT,
    request_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX idx_security_audit_logs_webhook_id ON security_audit_logs(webhook_id);
CREATE INDEX idx_security_audit_logs_created_at ON security_audit_logs(created_at);
CREATE INDEX idx_security_audit_logs_event_type ON security_audit_logs(event_type);
CREATE INDEX idx_security_audit_logs_client_ip ON security_audit_logs(client_ip);