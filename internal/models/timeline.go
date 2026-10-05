package models

import (
	"time"
)

type TimelineGroups struct {
	Overdue   []Item `json:"overdue"`
	Today     []Item `json:"today"`
	Tomorrow  []Item `json:"tomorrow"`
	Upcoming  []Item `json:"upcoming"`
	Completed []Item `json:"completed"`
}

// GroupItems groups items into timeline buckets based on local date/time.
func GroupItems(items []Item, now time.Time) TimelineGroups {
	groups := TimelineGroups{
		Overdue:   make([]Item, 0),
		Today:     make([]Item, 0),
		Tomorrow:  make([]Item, 0),
		Upcoming:  make([]Item, 0),
		Completed: make([]Item, 0),
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	dayAfterStart := todayStart.AddDate(0, 0, 2)

	for _, item := range items {
		if item.IsCompleted {
			groups.Completed = append(groups.Completed, item)
			continue
		}

		if item.StartAt.Before(todayStart) {
			if item.Type == TypeCollegeTask || item.Type == TypeWorkDue {
				groups.Overdue = append(groups.Overdue, item)
			}
		} else if item.StartAt.Before(tomorrowStart) {
			groups.Today = append(groups.Today, item)
		} else if item.StartAt.Before(dayAfterStart) {
			groups.Tomorrow = append(groups.Tomorrow, item)
		} else {
			groups.Upcoming = append(groups.Upcoming, item)
		}
	}

	return groups
}
