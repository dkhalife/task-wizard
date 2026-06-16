package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"taskwiz.app/core/internal/models"
	"taskwiz.app/core/internal/services/logging"
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

	endpoint := strings.TrimRight(provider.URL, "/") + "/message"
	parsedURL, err := validateOutboundURL(endpoint)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err.Error())
	}

	req, err := http.NewRequestWithContext(c, http.MethodPost, parsedURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", provider.Token)

	client := newSafeHTTPClient(10 * time.Second)
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
