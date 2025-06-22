-- Create job queue table for asynchronous webhook processing
CREATE TABLE IF NOT EXISTS job_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id TEXT NOT NULL REFERENCES webhooks(id),
    api_key_id INTEGER REFERENCES api_keys(id),
    
    -- Request data
    prompt TEXT NOT NULL,
    context TEXT,
    claude_options TEXT, -- JSON
    
    -- Queue management
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    priority INTEGER NOT NULL DEFAULT 0, -- Higher number = higher priority
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    
    -- Processing metadata
    worker_id TEXT,
    visibility_timeout TIMESTAMP, -- When the job becomes visible again if not completed
    error_message TEXT,
    
    -- Response data (stored after completion)
    response TEXT,
    execution_time_ms INTEGER,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes for efficient queue operations
CREATE INDEX idx_job_queue_status_priority ON job_queue(status, priority DESC, created_at);
CREATE INDEX idx_job_queue_visibility ON job_queue(status, visibility_timeout);
CREATE INDEX idx_job_queue_webhook_id ON job_queue(webhook_id);
CREATE INDEX idx_job_queue_created_at ON job_queue(created_at);

-- Enable WAL mode for better concurrency
PRAGMA journal_mode = WAL;