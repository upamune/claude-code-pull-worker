-- name: GetWebhook :one
SELECT * FROM webhooks WHERE id = ? AND is_active = 1;

-- name: ListWebhooks :many
SELECT * FROM webhooks WHERE is_active = 1 ORDER BY created_at DESC;

-- name: CreateWebhook :one
INSERT INTO webhooks (id, name, description, claude_options, notification_config)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateWebhook :exec
UPDATE webhooks 
SET name = ?, description = ?, claude_options = ?, notification_config = ?, updated_at = CURRENT_TIMESTAMP
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