package telemetry

import (
	"context"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"dkhalife.com/tasks/core/internal/version"
)

const (
	ServiceName = "task-wizard-api"
	dntKey      = "telemetry_dnt"
)

type dntKeyType struct{}

var dntContextKey = dntKeyType{}

func SetDNT(ctx context.Context) context.Context {
	return context.WithValue(ctx, dntContextKey, true)
}

func IsDNT(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	dnt, ok := ctx.Value(dntContextKey).(bool)
	return ok && dnt
}

func TrackEvent(ctx context.Context, name string, component string, properties map[string]string) {
	if IsDNT(ctx) {
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

func TrackError(ctx context.Context, name string, component string, err error, properties map[string]string) {
	if IsDNT(ctx) {
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

func TrackWarning(ctx context.Context, name string, component string, message string, properties map[string]string) {
	if IsDNT(ctx) {
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
