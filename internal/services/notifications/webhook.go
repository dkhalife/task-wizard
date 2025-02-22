package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	"dkhalife.com/tasks/core/internal/services/logging"
)

func SendNotificationViaWebhook(c context.Context, provider models.NotificationProvider, message string) error {
	log := logging.FromContext(c)

	// Create the payload
	payload := map[string]string{
		"message": message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error marshalling payload", err)
		return err
	}

	req, err := http.NewRequest(provider.Method, provider.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Error("Error creating HTTP request", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Received non-OK response", "status", resp.StatusCode)
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	return nil
}
