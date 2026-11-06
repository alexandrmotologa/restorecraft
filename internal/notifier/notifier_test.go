package notifier

import (
	"context"
	"testing"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

func TestNotifierDispatcherDryRun(t *testing.T) {
	cfg := config.NotificationConfig{
		Telegram: config.TelegramNotification{
			Enabled:  false,
			BotToken: "mock-token",
			ChatID:   "12345",
			NotifyOn: "all",
		},
		Slack: config.SlackNotification{
			Enabled:    false,
			WebhookURL: "https://hooks.slack.com/services/mock",
			NotifyOn:   "all",
		},
		Discord: config.DiscordNotification{
			Enabled:    false,
			WebhookURL: "https://discord.com/api/webhooks/mock",
			NotifyOn:   "all",
		},
	}

	dispatcher := NewDispatcher(cfg)
	if dispatcher == nil {
		t.Fatalf("expected NewDispatcher to return dispatcher")
	}

	drill := &models.DrillRun{
		ID:          "notifier-test-01",
		Name:        "Test Drill",
		Engine:      "sqlite",
		Status:      models.StatusCompleted,
		StartedAt:   time.Now(),
		HealthScore: 100,
		HealthGrade: models.GradeAPlus,
		SlaMet:      true,
	}

	// Dispatch without panic
	dispatcher.Dispatch(context.Background(), drill)

	summary := formatDrillSummary(drill)
	if summary == "" {
		t.Errorf("expected non-empty summary")
	}
}
