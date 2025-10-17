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

func TestGenerateVTODO_WithDueDate(t *testing.T) {
	dueDate := time.Date(2023, 3, 15, 14, 30, 0, 0, time.UTC)
	task := &models.Task{
		ID:          4,
		Title:       "Task with due date",
		CreatedAt:   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
		NextDueDate: &dueDate,
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Task with due date")
	require.Contains(t, vtodo, "DUE:20230315T143000Z")
}

func TestGenerateVTODO_WithoutDueDate(t *testing.T) {
	task := &models.Task{
		ID:          5,
		Title:       "Task without due date",
		CreatedAt:   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
		NextDueDate: nil,
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Task without due date")
	require.NotContains(t, vtodo, "DUE:")
}

func TestGenerateVTODO_DueDateUpdated(t *testing.T) {
	// Test that updating a due date generates correct VTODO
	dueDate := time.Date(2025, 10, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC)
	
	task := &models.Task{
		ID:          6,
		Title:       "Task with updated due date",
		CreatedAt:   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt:   &updatedAt,
		NextDueDate: &dueDate,
	}

	vtodo := generateVTODO(task)

	require.Contains(t, vtodo, "SUMMARY:Task with updated due date")
	require.Contains(t, vtodo, "DUE:20251010T100000Z")
	require.Contains(t, vtodo, "LAST-MODIFIED:20230201T120000Z")
}
