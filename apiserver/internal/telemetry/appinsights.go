package telemetry

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"dkhalife.com/tasks/core/internal/version"
)

var (
	client     appinsights.TelemetryClient
	clientOnce sync.Once
)

func GetAppInsightsClient() appinsights.TelemetryClient {
	clientOnce.Do(func() {
		connectionString := os.Getenv("APPINSIGHTS_CONNECTION_STRING")
		if connectionString == "" {
			log.Println("APPINSIGHTS_CONNECTION_STRING not set, Application Insights logging disabled")
			client = nil
			return
		}

		instrumentationKey, err := extractInstrumentationKey(connectionString)
		if err != nil {
			log.Printf("Error initializing Application Insights: failed to parse connection string: %v", err)
			client = nil
			return
		}

		telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)

		if endpoint, err := extractIngestionEndpoint(connectionString); err == nil && endpoint != "" {
			telemetryConfig.EndpointUrl = strings.TrimSuffix(endpoint, "/") + "/v2/track"
		}

		client = appinsights.NewTelemetryClientFromConfig(telemetryConfig)
		client.Context().Tags.Cloud().SetRole("task-wizard-api")
		client.Context().Tags.Application().SetVer(version.GetVersion())

		log.Printf("Application Insights initialized successfully (key: %s..., version: %s)", instrumentationKey[:8], version.GetFullVersion())
	})

	return client
}

func extractInstrumentationKey(connectionString string) (string, error) {
	return extractConnectionStringValue(connectionString, "InstrumentationKey")
}

func extractIngestionEndpoint(connectionString string) (string, error) {
	return extractConnectionStringValue(connectionString, "IngestionEndpoint")
}

func extractConnectionStringValue(connectionString, key string) (string, error) {
	pairs := splitConnectionString(connectionString)
	for _, pair := range pairs {
		k, v := splitKeyValue(pair)
		if k == key {
			return v, nil
		}
	}
	return "", fmt.Errorf("%s not found in connection string", key)
}

func splitConnectionString(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ';' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func splitKeyValue(s string) (string, string) {
	for i, c := range s {
		if c == '=' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func IsAppInsightsEnabled() bool {
	return GetAppInsightsClient() != nil
}

func FlushAppInsights() {
	if client := GetAppInsightsClient(); client != nil {
		<-client.Channel().Close(10 * time.Second)
		log.Println("Application Insights telemetry flushed successfully")
	}
}
