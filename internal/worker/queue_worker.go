package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/upamune/claude-code-pull-worker/internal/db"
	"github.com/upamune/claude-code-pull-worker/internal/executor"
	"github.com/upamune/claude-code-pull-worker/internal/models"
	"github.com/upamune/claude-code-pull-worker/internal/notifier"
	"github.com/upamune/claude-code-pull-worker/internal/notifier/discord"
)

type QueueWorker struct {
	id       string
	queries  *db.Queries
	executor *executor.ClaudeExecutor
	stopCh   chan struct{}
}

func NewQueueWorker(queries *db.Queries) *QueueWorker {
	return &QueueWorker{
		id:       uuid.New().String(),
		queries:  queries,
		executor: executor.NewClaudeExecutor(1 * time.Hour, queries),
		stopCh:   make(chan struct{}),
	}
}

func (w *QueueWorker) Start(ctx context.Context) {
	log.Printf("Queue worker %s started", w.id)
	
	// Reset stale jobs on startup
	if err := w.queries.ResetStaleJobs(ctx); err != nil {
		log.Printf("Failed to reset stale jobs: %v", err)
	}
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Queue worker %s stopping due to context cancellation", w.id)
			return
		case <-w.stopCh:
			log.Printf("Queue worker %s stopping", w.id)
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

func (w *QueueWorker) Stop() {
	close(w.stopCh)
}

func (w *QueueWorker) processNextJob(ctx context.Context) {
	// Try to dequeue a job
	job, err := w.queries.DequeueJob(ctx, sql.NullString{String: w.id, Valid: true})
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Failed to dequeue job: %v", err)
		}
		return
	}
	
	log.Printf("Worker %s processing job %d for webhook %s", w.id, job.ID, job.WebhookID)
	
	// Process the job
	startTime := time.Now()
	err = w.processJob(ctx, &job)
	executionTime := time.Since(startTime)
	
	if err != nil {
		log.Printf("Job %d failed: %v", job.ID, err)
		if err := w.queries.FailJob(ctx, db.FailJobParams{
			ID:           job.ID,
			ErrorMessage: sql.NullString{String: err.Error(), Valid: true},
		}); err != nil {
			log.Printf("Failed to mark job as failed: %v", err)
		}
		
		// Send failure notification
		go w.sendJobNotification(context.Background(), &job, nil, err, executionTime)
	}
}

func (w *QueueWorker) processJob(ctx context.Context, job *db.JobQueue) error {
	// Get webhook details
	_, err := w.queries.GetWebhook(ctx, job.WebhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}
	
	// Execute Claude with job options
	output, err := w.executor.ExecuteWithOptions(ctx, job.Prompt, *job)
	if err != nil {
		return fmt.Errorf("Claude execution failed: %w", err)
	}
	
	// Mark job as completed
	executionTimeMs := time.Since(job.StartedAt.Time).Milliseconds()
	if err := w.queries.CompleteJob(ctx, db.CompleteJobParams{
		ID:              job.ID,
		Response:        sql.NullString{String: output, Valid: true},
		ExecutionTimeMs: sql.NullInt64{Int64: executionTimeMs, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}
	
	// Also create execution history for backward compatibility
	_, err = w.queries.CreateExecutionHistory(ctx, db.CreateExecutionHistoryParams{
		WebhookID:       job.WebhookID,
		ApiKeyID:        job.ApiKeyID,
		Prompt:          job.Prompt,
		Response:        sql.NullString{String: output, Valid: true},
		Error:           sql.NullString{},
		Success:         true,
		ExecutionTimeMs: sql.NullInt64{Int64: executionTimeMs, Valid: true},
	})
	if err != nil {
		log.Printf("Failed to create execution history: %v", err)
	}
	
	// Send success notification
	go w.sendJobNotification(context.Background(), job, &output, nil, time.Duration(executionTimeMs)*time.Millisecond)
	
	log.Printf("Job %d completed successfully", job.ID)
	return nil
}


func (w *QueueWorker) sendJobNotification(ctx context.Context, job *db.JobQueue, response *string, err error, executionTime time.Duration) {
	// Get webhook for notification config
	webhook, webhookErr := w.queries.GetWebhook(ctx, job.WebhookID)
	if webhookErr != nil {
		log.Printf("Failed to get webhook for notification: %v", webhookErr)
		return
	}
	
	// Create webhook response object
	webhookResponse := models.NewWebhookResponse(job.Prompt, err == nil)
	webhookResponse.ExecutionTime = fmt.Sprintf("%.2fs", executionTime.Seconds())
	
	if err != nil {
		webhookResponse.Error = err.Error()
	} else if response != nil {
		webhookResponse.Response = *response
	}
	
	// Send notifications (reuse existing notification logic)
	w.sendNotifications(ctx, &webhook, webhookResponse)
}

func (w *QueueWorker) sendNotifications(ctx context.Context, webhook *db.Webhook, response *models.WebhookResponse) {
	// Parse notification config
	var notifConfig map[string]interface{}
	if notifBytes, ok := webhook.NotificationConfig.([]byte); ok {
		if err := json.Unmarshal(notifBytes, &notifConfig); err != nil {
			return
		}
	} else if notifStr, ok := webhook.NotificationConfig.(string); ok {
		if err := json.Unmarshal([]byte(notifStr), &notifConfig); err != nil {
			return
		}
	}
	
	// Build notifiers based on config
	var notifiers []notifier.Notifier
	
	// Check for Discord config
	if discordConfig, ok := notifConfig["discord"].(map[string]interface{}); ok {
		if webhookURL, ok := discordConfig["webhook_url"].(string); ok && webhookURL != "" {
			notifiers = append(notifiers, discord.NewClient(webhookURL))
		}
	}
	
	// If no webhook-specific config, check global settings
	if len(notifiers) == 0 {
		globalNotif, err := w.queries.GetGlobalSetting(ctx, "default_notification_config")
		if err == nil {
			var globalConfig map[string]interface{}
			if globalBytes, ok := globalNotif.([]byte); ok {
				if err := json.Unmarshal(globalBytes, &globalConfig); err == nil {
					if discordConfig, ok := globalConfig["discord"].(map[string]interface{}); ok {
						if webhookURL, ok := discordConfig["webhook_url"].(string); ok && webhookURL != "" {
							notifiers = append(notifiers, discord.NewClient(webhookURL))
						}
					}
				}
			} else if globalStr, ok := globalNotif.(string); ok {
				if err := json.Unmarshal([]byte(globalStr), &globalConfig); err == nil {
					if discordConfig, ok := globalConfig["discord"].(map[string]interface{}); ok {
						if webhookURL, ok := discordConfig["webhook_url"].(string); ok && webhookURL != "" {
							notifiers = append(notifiers, discord.NewClient(webhookURL))
						}
					}
				}
			}
		}
	}
	
	// Send notifications
	if len(notifiers) > 0 {
		multiNotifier := notifier.NewMultiNotifier(notifiers...)
		multiNotifier.SendNotification(response)
	}
}