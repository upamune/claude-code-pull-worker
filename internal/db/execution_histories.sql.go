// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: execution_histories.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const countExecutionHistoriesByWebhook = `-- name: CountExecutionHistoriesByWebhook :one
SELECT COUNT(*) as count FROM execution_histories WHERE webhook_id = ?
`

func (q *Queries) CountExecutionHistoriesByWebhook(ctx context.Context, webhookID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, countExecutionHistoriesByWebhook, webhookID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createExecutionHistory = `-- name: CreateExecutionHistory :one
INSERT INTO execution_histories (webhook_id, api_key_id, prompt, response, error, success, execution_time_ms)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id, webhook_id, api_key_id, prompt, response, error, success, execution_time_ms, created_at
`

type CreateExecutionHistoryParams struct {
	WebhookID       string         `json:"webhook_id"`
	ApiKeyID        sql.NullInt64  `json:"api_key_id"`
	Prompt          string         `json:"prompt"`
	Response        sql.NullString `json:"response"`
	Error           sql.NullString `json:"error"`
	Success         bool           `json:"success"`
	ExecutionTimeMs sql.NullInt64  `json:"execution_time_ms"`
}

func (q *Queries) CreateExecutionHistory(ctx context.Context, arg CreateExecutionHistoryParams) (ExecutionHistory, error) {
	row := q.db.QueryRowContext(ctx, createExecutionHistory,
		arg.WebhookID,
		arg.ApiKeyID,
		arg.Prompt,
		arg.Response,
		arg.Error,
		arg.Success,
		arg.ExecutionTimeMs,
	)
	var i ExecutionHistory
	err := row.Scan(
		&i.ID,
		&i.WebhookID,
		&i.ApiKeyID,
		&i.Prompt,
		&i.Response,
		&i.Error,
		&i.Success,
		&i.ExecutionTimeMs,
		&i.CreatedAt,
	)
	return i, err
}

const getExecutionHistory = `-- name: GetExecutionHistory :one
SELECT id, webhook_id, api_key_id, prompt, response, error, success, execution_time_ms, created_at FROM execution_histories WHERE id = ?
`

func (q *Queries) GetExecutionHistory(ctx context.Context, id int64) (ExecutionHistory, error) {
	row := q.db.QueryRowContext(ctx, getExecutionHistory, id)
	var i ExecutionHistory
	err := row.Scan(
		&i.ID,
		&i.WebhookID,
		&i.ApiKeyID,
		&i.Prompt,
		&i.Response,
		&i.Error,
		&i.Success,
		&i.ExecutionTimeMs,
		&i.CreatedAt,
	)
	return i, err
}

const getExecutionStats = `-- name: GetExecutionStats :one
SELECT 
    COUNT(*) as total_executions,
    SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_executions,
    AVG(execution_time_ms) as avg_execution_time_ms
FROM execution_histories
WHERE webhook_id = ? AND created_at >= ?
`

type GetExecutionStatsParams struct {
	WebhookID string    `json:"webhook_id"`
	CreatedAt time.Time `json:"created_at"`
}

type GetExecutionStatsRow struct {
	TotalExecutions      int64           `json:"total_executions"`
	SuccessfulExecutions sql.NullFloat64 `json:"successful_executions"`
	AvgExecutionTimeMs   sql.NullFloat64 `json:"avg_execution_time_ms"`
}

func (q *Queries) GetExecutionStats(ctx context.Context, arg GetExecutionStatsParams) (GetExecutionStatsRow, error) {
	row := q.db.QueryRowContext(ctx, getExecutionStats, arg.WebhookID, arg.CreatedAt)
	var i GetExecutionStatsRow
	err := row.Scan(&i.TotalExecutions, &i.SuccessfulExecutions, &i.AvgExecutionTimeMs)
	return i, err
}

const getLastExecution = `-- name: GetLastExecution :one
SELECT id, webhook_id, api_key_id, prompt, response, error, success, execution_time_ms, created_at FROM execution_histories 
WHERE webhook_id = ? AND success = 1
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetLastExecution(ctx context.Context, webhookID string) (ExecutionHistory, error) {
	row := q.db.QueryRowContext(ctx, getLastExecution, webhookID)
	var i ExecutionHistory
	err := row.Scan(
		&i.ID,
		&i.WebhookID,
		&i.ApiKeyID,
		&i.Prompt,
		&i.Response,
		&i.Error,
		&i.Success,
		&i.ExecutionTimeMs,
		&i.CreatedAt,
	)
	return i, err
}

const listExecutionHistoriesByWebhook = `-- name: ListExecutionHistoriesByWebhook :many
SELECT id, webhook_id, api_key_id, prompt, response, error, success, execution_time_ms, created_at FROM execution_histories 
WHERE webhook_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?
`

type ListExecutionHistoriesByWebhookParams struct {
	WebhookID string `json:"webhook_id"`
	Limit     int64  `json:"limit"`
	Offset    int64  `json:"offset"`
}

func (q *Queries) ListExecutionHistoriesByWebhook(ctx context.Context, arg ListExecutionHistoriesByWebhookParams) ([]ExecutionHistory, error) {
	rows, err := q.db.QueryContext(ctx, listExecutionHistoriesByWebhook, arg.WebhookID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ExecutionHistory{}
	for rows.Next() {
		var i ExecutionHistory
		if err := rows.Scan(
			&i.ID,
			&i.WebhookID,
			&i.ApiKeyID,
			&i.Prompt,
			&i.Response,
			&i.Error,
			&i.Success,
			&i.ExecutionTimeMs,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
