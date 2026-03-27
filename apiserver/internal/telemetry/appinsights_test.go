package telemetry

import (
	"testing"
)

func TestExtractInstrumentationKey(t *testing.T) {
	key, err := extractInstrumentationKey("InstrumentationKey=abc12345-def6-7890-abcd-ef1234567890;IngestionEndpoint=https://eastus-8.in.applicationinsights.azure.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "abc12345-def6-7890-abcd-ef1234567890" {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestExtractInstrumentationKeyMissing(t *testing.T) {
	_, err := extractInstrumentationKey("IngestionEndpoint=https://example.com/")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestExtractIngestionEndpoint(t *testing.T) {
	endpoint, err := extractIngestionEndpoint("InstrumentationKey=abc;IngestionEndpoint=https://eastus-8.in.applicationinsights.azure.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "https://eastus-8.in.applicationinsights.azure.com/" {
		t.Fatalf("unexpected endpoint: %s", endpoint)
	}
}

func TestSplitConnectionString(t *testing.T) {
	parts := splitConnectionString("Key1=Value1;Key2=Value2;Key3=Value3")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	if parts[0] != "Key1=Value1" || parts[1] != "Key2=Value2" || parts[2] != "Key3=Value3" {
		t.Fatalf("unexpected parts: %v", parts)
	}
}

func TestSplitKeyValue(t *testing.T) {
	k, v := splitKeyValue("InstrumentationKey=abc123")
	if k != "InstrumentationKey" || v != "abc123" {
		t.Fatalf("unexpected key/value: %s=%s", k, v)
	}
}

func TestSplitKeyValueNoEquals(t *testing.T) {
	k, v := splitKeyValue("noequals")
	if k != "noequals" || v != "" {
		t.Fatalf("unexpected key/value: %s=%s", k, v)
	}
}

func TestIsAppInsightsEnabledWithoutEnvVar(t *testing.T) {
	// Note: due to sync.Once, the singleton may already be initialized from other tests.
	// In a fresh process with no env var, IsAppInsightsEnabled returns false.
	// This test at minimum verifies the function doesn't panic.
	_ = IsAppInsightsEnabled()
}
