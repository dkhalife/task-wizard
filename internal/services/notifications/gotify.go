package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"dkhalife.com/tasks/core/internal/models"
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

	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", provider.URL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", provider.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send HTTP request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Gotify returned non-OK status: %v", resp.Status)
		return fmt.Errorf("Gotify returned non-OK status: %v", resp.Status)
	}

	log.Printf("Notification sent successfully via Gotify")
	return nil
}
