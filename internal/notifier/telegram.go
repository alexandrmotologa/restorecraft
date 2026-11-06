package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// SendTelegram posts drill summary to a Telegram chat.
func SendTelegram(ctx context.Context, botToken string, chatID string, drill *models.DrillRun) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	text := formatDrillSummary(drill)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram api error: status %d", resp.StatusCode)
	}

	return nil
}
