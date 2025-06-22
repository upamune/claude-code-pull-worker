package notifier

import "github.com/upamune/claude-code-pull-worker/internal/models"

// Notifier defines the interface for sending notifications
type Notifier interface {
	SendNotification(response *models.WebhookResponse) error
	Name() string
}

// MultiNotifier allows sending notifications to multiple services
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier creates a new MultiNotifier
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{
		notifiers: notifiers,
	}
}

// SendNotification sends notifications to all configured notifiers
func (m *MultiNotifier) SendNotification(response *models.WebhookResponse) error {
	for _, n := range m.notifiers {
		// Continue sending to other notifiers even if one fails
		go n.SendNotification(response)
	}
	return nil
}

// Name returns the name of the multi-notifier
func (m *MultiNotifier) Name() string {
	return "multi-notifier"
}