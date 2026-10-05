package models_test

import (
	"testing"
	"time"
	"todo/internal/models"
)

func TestItem_Validate(t *testing.T) {
	now := time.Now()
	later := now.Add(2 * time.Hour)
	earlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name    string
		item    models.Item
		wantErr bool
	}{
		{
			name: "valid college task",
			item: models.Item{
				Title:    "Math Assignment",
				Type:     models.TypeCollegeTask,
				StartAt:  now,
				Priority: models.PriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "empty title fails",
			item: models.Item{
				Title:   "   ",
				Type:    models.TypeWorkDue,
				StartAt: now,
			},
			wantErr: true,
		},
		{
			name: "invalid type fails",
			item: models.Item{
				Title:   "Invalid",
				Type:    "unknown_type",
				StartAt: now,
			},
			wantErr: true,
		},
		{
			name: "event EndAt earlier than StartAt fails",
			item: models.Item{
				Title:    "Lecture",
				Type:     models.TypeEvent,
				StartAt:  now,
				EndAt:    &earlier,
				Priority: models.PriorityLow,
			},
			wantErr: true,
		},
		{
			name: "event EndAt after StartAt passes",
			item: models.Item{
				Title:    "Lecture",
				Type:     models.TypeEvent,
				StartAt:  now,
				EndAt:    &later,
				Priority: models.PriorityLow,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.item.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
