package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"taskwiz.app/core/internal/models"
)

func SendNotificationViaWebhook(c context.Context, provider models.NotificationProvider, message string) error {
	parsedURL, err := validateOutboundURL(provider.URL)
	if err != nil {
		return err
	}

	method, err := validateOutboundMethod(provider.Method)
	if err != nil {
		return err
	}

	payload := map[string]string{
		"message": message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %s", err.Error())
	}

	req, err := http.NewRequestWithContext(c, method, parsedURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := newSafeHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %s", err.Error())
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %d", resp.StatusCode)
	}

	return nil
}
