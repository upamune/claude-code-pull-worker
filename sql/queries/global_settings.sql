-- name: GetGlobalSetting :one
SELECT setting_value FROM global_settings WHERE setting_key = ?;

-- name: UpdateGlobalSetting :exec
UPDATE global_settings 
SET setting_value = ?, updated_at = CURRENT_TIMESTAMP
WHERE setting_key = ?;

-- name: ListGlobalSettings :many
SELECT * FROM global_settings ORDER BY setting_key;