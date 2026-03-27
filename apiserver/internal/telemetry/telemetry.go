package telemetry

import (
	"github.com/gin-gonic/gin"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"dkhalife.com/tasks/core/internal/version"
)

const (
	ServiceName = "task-wizard-api"
	dntKey      = "telemetry_dnt"
)

func SetDNT(c *gin.Context) {
	c.Set(dntKey, true)
}

func IsDNT(c *gin.Context) bool {
	v, exists := c.Get(dntKey)
	if !exists {
		return false
	}
	dnt, ok := v.(bool)
	return ok && dnt
}

func TrackEvent(c *gin.Context, name string, component string, properties map[string]string) {
	if c != nil && IsDNT(c) {
		return
	}

	client := GetAppInsightsClient()
	if client == nil {
		return
	}

	event := appinsights.NewEventTelemetry(name)
	event.Properties["service_name"] = ServiceName
	event.Properties["service_version"] = version.GetVersion()
	event.Properties["build_number"] = version.BuildNumber
	event.Properties["commit_hash"] = version.CommitHash
	event.Properties["app_component"] = component

	for k, v := range properties {
		event.Properties[k] = v
	}

	client.Track(event)
}

func TrackError(c *gin.Context, name string, component string, err error, properties map[string]string) {
	if c != nil && IsDNT(c) {
		return
	}

	client := GetAppInsightsClient()
	if client == nil {
		return
	}

	event := appinsights.NewEventTelemetry(name)
	event.Properties["service_name"] = ServiceName
	event.Properties["service_version"] = version.GetVersion()
	event.Properties["build_number"] = version.BuildNumber
	event.Properties["commit_hash"] = version.CommitHash
	event.Properties["app_component"] = component
	event.Properties["log_level"] = "error"

	if err != nil {
		event.Properties["error"] = err.Error()
	}

	for k, v := range properties {
		event.Properties[k] = v
	}

	client.Track(event)
}

func TrackWarning(c *gin.Context, name string, component string, message string, properties map[string]string) {
	if c != nil && IsDNT(c) {
		return
	}

	client := GetAppInsightsClient()
	if client == nil {
		return
	}

	event := appinsights.NewEventTelemetry(name)
	event.Properties["service_name"] = ServiceName
	event.Properties["service_version"] = version.GetVersion()
	event.Properties["build_number"] = version.BuildNumber
	event.Properties["commit_hash"] = version.CommitHash
	event.Properties["app_component"] = component
	event.Properties["log_level"] = "warn"
	event.Properties["message"] = message

	for k, v := range properties {
		event.Properties[k] = v
	}

	client.Track(event)
}
