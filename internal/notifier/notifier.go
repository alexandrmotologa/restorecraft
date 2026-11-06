package notifier

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// Dispatcher coordinates multi-channel alerts for drill outcomes.
type Dispatcher struct {
	cfg config.NotificationConfig
}

// NewDispatcher creates a notification dispatcher from configuration.
func NewDispatcher(cfg config.NotificationConfig) *Dispatcher {
	return &Dispatcher{cfg: cfg}
}

// Dispatch sends notifications for a completed or failed drill to all enabled channels.
func (d *Dispatcher) Dispatch(ctx context.Context, drill *models.DrillRun) {
	isFailure := drill.Status == models.StatusFailed || drill.HealthGrade == models.GradeF || !drill.SlaMet

	// 1. Telegram
	if d.cfg.Telegram.Enabled && d.cfg.Telegram.BotToken != "" && d.cfg.Telegram.ChatID != "" {
		if d.cfg.Telegram.NotifyOn == "all" || (d.cfg.Telegram.NotifyOn == "failures_only" && isFailure) {
			go func() {
				_ = SendTelegram(ctx, d.cfg.Telegram.BotToken, d.cfg.Telegram.ChatID, drill)
			}()
		}
	}

	// 2. Slack
	if d.cfg.Slack.Enabled && d.cfg.Slack.WebhookURL != "" {
		if d.cfg.Slack.NotifyOn == "all" || (d.cfg.Slack.NotifyOn == "failures_only" && isFailure) {
			go func() {
				_ = SendSlack(ctx, d.cfg.Slack.WebhookURL, drill)
			}()
		}
	}

	// 3. Discord
	if d.cfg.Discord.Enabled && d.cfg.Discord.WebhookURL != "" {
		if d.cfg.Discord.NotifyOn == "all" || (d.cfg.Discord.NotifyOn == "failures_only" && isFailure) {
			go func() {
				_ = SendDiscord(ctx, d.cfg.Discord.WebhookURL, drill)
			}()
		}
	}
}

func formatDrillSummary(drill *models.DrillRun) string {
	icon := "✅"
	if drill.Status == models.StatusFailed || drill.HealthGrade == models.GradeF {
		icon = "🚨"
	} else if drill.HealthGrade == models.GradeB || drill.HealthGrade == models.GradeC {
		icon = "⚠️"
	}

	slaText := "Met"
	if !drill.SlaMet {
		slaText = "EXCEEDED"
	}

	msg := fmt.Sprintf("%s restorecraft Drill: %s\nEngine: %s\nStatus: %s\nHealth Grade: %s (Score: %.1f)\nTotal RTO: %dms (SLA: %ds, %s)\nIntegrity: %t",
		icon,
		drill.Name,
		drill.Engine,
		drill.Status,
		drill.HealthGrade,
		drill.HealthScore,
		drill.Timings.TotalRtoMs,
		drill.RtoSlaSeconds,
		slaText,
		drill.Audit.IntegrityOk,
	)

	if drill.ErrorMessage != "" {
		msg += fmt.Sprintf("\nError: %s", drill.ErrorMessage)
	}

	return msg
}
