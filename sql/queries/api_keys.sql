-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys 
WHERE key_hash = ? AND is_active = 1;

-- name: GetAPIKeysForWebhook :many
SELECT * FROM api_keys
WHERE webhook_id = ? AND is_active = 1;

-- name: ListAPIKeysByWebhook :many
SELECT id, webhook_id, key_prefix, key_suffix, description, created_at, last_used_at
FROM api_keys 
WHERE webhook_id = ? AND is_active = 1
ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (webhook_id, key_hash, key_prefix, key_suffix, description)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteAPIKey :exec
UPDATE api_keys SET is_active = 0 WHERE id = ?;

-- name: GetAPIKeyWithWebhook :one
SELECT 
    ak.*,
    w.name as webhook_name,
    w.claude_options,
    w.notification_config
FROM api_keys ak
JOIN webhooks w ON ak.webhook_id = w.id
WHERE ak.key_hash = ? AND ak.is_active = 1 AND w.is_active = 1;