package caldav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseVTODO_WithDueDateTimestamp(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Test Task
DUE:20240315T120000Z
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)
	require.NoError(t, err)
	require.Equal(t, "Test Task", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDue, *due)
}

func TestParseVTODO_WithDueDateOnly(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Test Task
DUE:20240315
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)
	require.NoError(t, err)
	require.Equal(t, "Test Task", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDue, *due)
}

func TestParseVTODO_WithoutDueDate(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Test Task Without Due Date
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)
	require.NoError(t, err)
	require.Equal(t, "Test Task Without Due Date", title)
	require.Nil(t, due)
}

func TestParseVTODO_WithInvalidDueDate(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Test Task
DUE:invalid-date-format
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)
	require.NoError(t, err)
	require.Equal(t, "Test Task", title)
	// Invalid date formats are gracefully ignored, returning nil due date to allow partial data extraction.
	require.Nil(t, due)
}

func TestParseVTODO_TitleAndDueDateTogether(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VTODO
SUMMARY:Important Task
DUE:20250101T000000Z
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)
	require.NoError(t, err)
	require.Equal(t, "Important Task", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDue, *due)
}

func TestParseVTODO_EmptyString(t *testing.T) {
	title, due, err := ParseVTODO("")
	require.NoError(t, err)
	require.Empty(t, title)
	require.Nil(t, due)
}
