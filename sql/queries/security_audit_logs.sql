-- name: LogSecurityAuditEvent :exec
INSERT INTO security_audit_logs (
    webhook_id,
    event_type,
    client_ip,
    user_agent,
    api_key_provided,
    error_message,
    request_path
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetSecurityAuditLogs :many
SELECT * FROM security_audit_logs
WHERE webhook_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetSecurityAuditLogsByType :many
SELECT * FROM security_audit_logs
WHERE webhook_id = ? AND event_type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountSecurityAuditEvents :one
SELECT COUNT(*) FROM security_audit_logs
WHERE webhook_id = ? AND created_at > ?;

-- name: GetRecentSecurityAuditLogs :many
SELECT * FROM security_audit_logs
ORDER BY created_at DESC
LIMIT ?;

-- name: GetSecurityAuditLogsByIP :many
SELECT * FROM security_audit_logs
WHERE client_ip = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;