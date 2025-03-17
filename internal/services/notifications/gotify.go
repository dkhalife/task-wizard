package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
)

type GotifyMessage struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

func SendNotificationViaGotify(c context.Context, provider models.NotificationProvider, message string) error {
	msg := GotifyMessage{
		Message: message,
	}

	log := logging.FromContext(c)
	log.Debug("Sending notification via Gotify")

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err.Error())
	}

	req, err := http.NewRequest("POST", provider.URL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", provider.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotify returned non-OK status: %d", resp.StatusCode)
	}

	log.Debug("Notification sent successfully via Gotify")
	return nil
}
