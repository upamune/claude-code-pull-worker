-- name: GetWebhook :one
SELECT * FROM webhooks WHERE id = ? AND is_active = 1;

-- name: ListWebhooks :many
SELECT * FROM webhooks WHERE is_active = 1 ORDER BY created_at DESC;

-- name: CreateWebhook :one
INSERT INTO webhooks (
    id, name, description, notification_config,
    working_dir, max_thinking_tokens, max_turns,
    custom_system_prompt, append_system_prompt,
    allowed_tools, disallowed_tools,
    permission_mode, permission_prompt_tool_name,
    model, fallback_model, mcp_servers
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateWebhook :exec
UPDATE webhooks 
SET name = ?, 
    description = ?, 
    notification_config = ?,
    working_dir = ?,
    max_thinking_tokens = ?,
    max_turns = ?,
    custom_system_prompt = ?,
    append_system_prompt = ?,
    allowed_tools = ?,
    disallowed_tools = ?,
    permission_mode = ?,
    permission_prompt_tool_name = ?,
    model = ?,
    fallback_model = ?,
    mcp_servers = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteWebhook :exec
UPDATE webhooks SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetWebhookWithStats :one
SELECT 
    w.*,
    COUNT(DISTINCT ak.id) as api_key_count,
    COUNT(DISTINCT eh.id) as execution_count,
    MAX(eh.created_at) as last_execution
FROM webhooks w
LEFT JOIN api_keys ak ON w.id = ak.webhook_id AND ak.is_active = 1
LEFT JOIN execution_histories eh ON w.id = eh.webhook_id
WHERE w.id = ? AND w.is_active = 1
GROUP BY w.id;