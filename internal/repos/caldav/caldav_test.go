package caldav

import (
	"testing"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	"github.com/stretchr/testify/require"
)

func TestGenerateVTODO_TitleNewline(t *testing.T) {
	task := &models.Task{
		ID:        1,
		Title:     "Line1\nLine2\rLine3",
		CreatedAt: time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Line1Line2Line3")
	require.NotContains(t, vtodo, "Line1\nLine2")
	require.NotContains(t, vtodo, "\r")
}

func TestGenerateVTODO_TitleSpecialChars(t *testing.T) {
	task := &models.Task{
		ID:        2,
		Title:     "Task;with,chars",
		CreatedAt: time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Task\\;with\\,chars")
}

func TestGenerateVTODO_TitleBackslash(t *testing.T) {
	task := &models.Task{
		ID:        3,
		Title:     "Back\\slash",
		CreatedAt: time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Back\\\\slash")
}
