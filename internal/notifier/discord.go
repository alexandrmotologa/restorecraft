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

// SendDiscord posts drill summary to a Discord webhook URL.
func SendDiscord(ctx context.Context, webhookURL string, drill *models.DrillRun) error {
	text := formatDrillSummary(drill)

	payload := map[string]string{
		"content": text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
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
		return fmt.Errorf("discord webhook error: status %d", resp.StatusCode)
	}

	return nil
}
