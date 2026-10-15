package models_test

import (
	"testing"
	"time"
	"todo/internal/models"
)

func TestGroupItems(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)
	overdue := models.Item{ID: 1, Title: "Overdue Task", Type: models.TypeTask, StartAt: now.Add(-24 * time.Hour), IsCompleted: false}
	todayItem := models.Item{ID: 2, Title: "Today Task", Type: models.TypeTask, StartAt: now.Add(2 * time.Hour), IsCompleted: false}
	tomorrowItem := models.Item{ID: 3, Title: "Tomorrow Event", Type: models.TypeEvent, StartAt: now.Add(26 * time.Hour)}
	upcomingItem := models.Item{ID: 4, Title: "Next Week Task", Type: models.TypeTask, StartAt: now.Add(5 * 24 * time.Hour)}
	completedToday := models.Item{ID: 5, Title: "Done Today", Type: models.TypeTask, StartAt: now.Add(1 * time.Hour), IsCompleted: true}
	completedPast := models.Item{ID: 6, Title: "Done Yesterday", Type: models.TypeTask, StartAt: now.Add(-24 * time.Hour), IsCompleted: true}

	groups := models.GroupItems([]models.Item{overdue, todayItem, tomorrowItem, upcomingItem, completedToday, completedPast}, now)

	if len(groups.Overdue) != 1 || groups.Overdue[0].ID != 1 {
		t.Errorf("expected 1 overdue item, got %d", len(groups.Overdue))
	}
	if len(groups.Today) != 1 || groups.Today[0].ID != 2 {
		t.Errorf("expected 1 today item, got %d", len(groups.Today))
	}
	if len(groups.Tomorrow) != 1 || groups.Tomorrow[0].ID != 3 {
		t.Errorf("expected 1 tomorrow item, got %d", len(groups.Tomorrow))
	}
	if len(groups.Upcoming) != 1 || groups.Upcoming[0].ID != 4 {
		t.Errorf("expected 1 upcoming item, got %d", len(groups.Upcoming))
	}
	// Completed past end/due date must be hidden, only completed today shown
	if len(groups.Completed) != 1 || groups.Completed[0].ID != 5 {
		t.Errorf("expected 1 completed item (id 5), got %d", len(groups.Completed))
	}
}
