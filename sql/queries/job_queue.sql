-- name: EnqueueJob :one
INSERT INTO job_queue (
    webhook_id,
    api_key_id,
    prompt,
    context,
    claude_options,
    priority
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: DequeueJob :one
UPDATE job_queue
SET 
    status = 'processing',
    started_at = CURRENT_TIMESTAMP,
    visibility_timeout = datetime('now', '+10 minutes'),
    worker_id = ?
WHERE id = (
    SELECT id FROM job_queue
    WHERE status = 'pending'
       OR (status = 'processing' AND visibility_timeout < CURRENT_TIMESTAMP)
    ORDER BY priority DESC, created_at ASC
    LIMIT 1
)
RETURNING *;

-- name: CompleteJob :exec
UPDATE job_queue
SET 
    status = 'completed',
    completed_at = CURRENT_TIMESTAMP,
    response = ?,
    execution_time_ms = ?
WHERE id = ? AND status = 'processing';

-- name: FailJob :exec
UPDATE job_queue
SET 
    status = CASE 
        WHEN retry_count >= max_retries THEN 'failed'
        ELSE 'pending'
    END,
    retry_count = retry_count + 1,
    error_message = ?,
    visibility_timeout = NULL,
    worker_id = NULL
WHERE id = ? AND status = 'processing';

-- name: ResetStaleJobs :exec
UPDATE job_queue
SET 
    status = 'pending',
    visibility_timeout = NULL,
    worker_id = NULL
WHERE status = 'processing' 
  AND visibility_timeout < CURRENT_TIMESTAMP;

-- name: GetJobStatus :one
SELECT * FROM job_queue WHERE id = ?;

-- name: GetPendingJobCount :one
SELECT COUNT(*) as count FROM job_queue WHERE status = 'pending';

-- name: GetRecentJobs :many
SELECT * FROM job_queue
ORDER BY created_at DESC
LIMIT ?;

-- name: GetJobsByWebhook :many
SELECT * FROM job_queue
WHERE webhook_id = ?
ORDER BY created_at DESC
LIMIT ?;