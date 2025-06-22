-- name: GetGlobalSetting :one
SELECT value FROM global_settings WHERE key = ?;

-- name: UpdateGlobalSetting :exec
UPDATE global_settings 
SET value = ?, updated_at = CURRENT_TIMESTAMP
WHERE key = ?;

-- name: ListGlobalSettings :many
SELECT * FROM global_settings ORDER BY key;