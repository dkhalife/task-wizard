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
DUE:20230115T120000Z
CATEGORIES:
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "Test Task", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)
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
UID:2
SUMMARY:Test Task Date Only
DUE:20230116
CATEGORIES:
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "Test Task Date Only", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC)
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
UID:3
SUMMARY:Test Task No Due Date
CATEGORIES:
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "Test Task No Due Date", title)
	require.Nil(t, due)
}

func TestParseVTODO_UpdateDueDate(t *testing.T) {
	// Simulate updating a task with a new due date
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Updated Task Title
DUE:20250315T140000Z
CATEGORIES:
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "Updated Task Title", title)
	require.NotNil(t, due)
	
	expectedDue := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDue, *due)
}

func TestParseVTODO_RemoveDueDate(t *testing.T) {
	// Simulate updating a task to remove the due date
	vtodo := `BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:20230102T030405Z
LAST-MODIFIED:20230102T030405Z
DTSTAMP:20230102T030405Z
UID:1
SUMMARY:Task Without Due Date
CATEGORIES:
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "Task Without Due Date", title)
	require.Nil(t, due)
}

func TestParseVTODO_EmptyVTODO(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VTODO
END:VTODO
END:VCALENDAR`

	title, due, err := ParseVTODO(vtodo)

	require.NoError(t, err)
	require.Equal(t, "", title)
	require.Nil(t, due)
}
