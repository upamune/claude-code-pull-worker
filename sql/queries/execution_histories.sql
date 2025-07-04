-- name: CreateExecutionHistory :one
INSERT INTO execution_histories (webhook_id, api_key_id, prompt, response, error, success, execution_time_ms)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListExecutionHistoriesByWebhook :many
SELECT * FROM execution_histories 
WHERE webhook_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetExecutionHistory :one
SELECT * FROM execution_histories WHERE id = ?;

-- name: CountExecutionHistoriesByWebhook :one
SELECT COUNT(*) as count FROM execution_histories WHERE webhook_id = ?;

-- name: GetExecutionStats :one
SELECT 
    COUNT(*) as total_executions,
    SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_executions,
    AVG(execution_time_ms) as avg_execution_time_ms
FROM execution_histories
WHERE webhook_id = ? AND created_at >= ?;

-- name: GetLastExecution :one
SELECT * FROM execution_histories 
WHERE webhook_id = ? AND success = 1
ORDER BY created_at DESC
LIMIT 1;