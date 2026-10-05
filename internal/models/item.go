package models

import (
	"errors"
	"strings"
	"time"
)

type ItemType string

const (
	TypeCollegeTask ItemType = "college_task"
	TypeWorkDue     ItemType = "work_due"
	TypeEvent       ItemType = "event"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Item struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Type          ItemType   `json:"type"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         *time.Time `json:"end_at,omitempty"`
	ContextTag    string     `json:"context_tag"`
	LocationOrURL string     `json:"location_or_url"`
	Priority      Priority   `json:"priority"`
	IsCompleted   bool       `json:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Validate validates the Item fields according to business rules.
func (i *Item) Validate() error {
	i.Title = strings.TrimSpace(i.Title)
	if i.Title == "" {
		return errors.New("title is required")
	}

	switch i.Type {
	case TypeCollegeTask, TypeWorkDue, TypeEvent:
		// Valid type
	default:
		return errors.New("invalid item type; must be college_task, work_due, or event")
	}

	if i.Priority == "" {
		i.Priority = PriorityMedium
	}
	switch i.Priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		// Valid priority
	default:
		return errors.New("invalid priority; must be low, medium, or high")
	}

	if i.EndAt != nil && i.StartAt.After(*i.EndAt) {
		return errors.New("end time cannot be earlier than start time")
	}

	return nil
}
