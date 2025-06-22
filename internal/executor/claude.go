package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ClaudeExecutor struct {
	timeout time.Duration
}

func NewClaudeExecutor(timeout time.Duration) *ClaudeExecutor {
	return &ClaudeExecutor{
		timeout: timeout,
	}
}

func (e *ClaudeExecutor) Execute(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timeout after %v", e.timeout)
		}
		return "", fmt.Errorf("execution error: %v\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}