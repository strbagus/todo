package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"todo/internal/models"
)

type ItemRepo struct {
	db *sql.DB
}

func NewItemRepo(db *sql.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

// Create inserts a new Item into SQLite and populates generated ID and timestamps.
func (r *ItemRepo) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (
			title, description, type, start_at, end_at,
			context_tag, location_or_url, priority, is_completed, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	isCompletedInt := 0
	if item.IsCompleted {
		isCompletedInt = 1
	}

	err := r.db.QueryRowContext(
		ctx, query,
		item.Title, item.Description, string(item.Type),
		item.StartAt.UTC().Format(time.RFC3339),
		formatNullableTime(item.EndAt),
		item.ContextTag, item.LocationOrURL, string(item.Priority),
		isCompletedInt, formatNullableTime(item.CompletedAt),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	return nil
}

// GetByID fetches a single Item by its primary key ID.
func (r *ItemRepo) GetByID(ctx context.Context, id int64) (*models.Item, error) {
	query := `
		SELECT id, title, description, type, start_at, end_at,
		       context_tag, location_or_url, priority, is_completed,
		       completed_at, created_at, updated_at
		FROM items
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanItem(row)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// List retrieves items filtered by type and optional search keyword, ordered by start_at ascending.
func (r *ItemRepo) List(ctx context.Context, filterType models.ItemType, search string) ([]models.Item, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT id, title, description, type, start_at, end_at,
		       context_tag, location_or_url, priority, is_completed,
		       completed_at, created_at, updated_at
		FROM items
		WHERE 1=1
	`)

	var args []any
	if filterType != "" {
		query.WriteString(" AND type = ?")
		args = append(args, string(filterType))
	}

	search = strings.TrimSpace(search)
	if search != "" {
		pattern := "%" + search + "%"
		query.WriteString(" AND (title LIKE ? OR context_tag LIKE ? OR description LIKE ?)")
		args = append(args, pattern, pattern, pattern)
	}

	query.WriteString(" ORDER BY start_at ASC")

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return items, nil
}

// ToggleCompleted toggles the is_completed status and sets/unsets completed_at.
func (r *ItemRepo) ToggleCompleted(ctx context.Context, id int64) (*models.Item, error) {
	query := `
		UPDATE items
		SET is_completed = 1 - is_completed,
		    completed_at = CASE WHEN is_completed = 0 THEN CURRENT_TIMESTAMP ELSE NULL END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, title, description, type, start_at, end_at,
		          context_tag, location_or_url, priority, is_completed,
		          completed_at, created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanItem(row)
}

// Update updates an existing Item in SQLite.
func (r *ItemRepo) Update(ctx context.Context, item *models.Item) error {
	query := `
		UPDATE items
		SET title = ?, description = ?, type = ?, start_at = ?, end_at = ?,
		    context_tag = ?, location_or_url = ?, priority = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	res, err := r.db.ExecContext(
		ctx, query,
		item.Title, item.Description, string(item.Type),
		item.StartAt.UTC().Format(time.RFC3339),
		formatNullableTime(item.EndAt),
		item.ContextTag, item.LocationOrURL, string(item.Priority),
		item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found with id %d", item.ID)
	}

	return nil
}

// Delete removes an item from SQLite by ID.
func (r *ItemRepo) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM items WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found with id %d", id)
	}

	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanItem(s scannable) (*models.Item, error) {
	var item models.Item
	var itemTypeStr, priorityStr string
	var startAtStr string
	var endAtStr, completedAtStr sql.NullString
	var isCompletedInt int

	err := s.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&itemTypeStr,
		&startAtStr,
		&endAtStr,
		&item.ContextTag,
		&item.LocationOrURL,
		&priorityStr,
		&isCompletedInt,
		&completedAtStr,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	item.Type = models.ItemType(itemTypeStr)
	item.Priority = models.Priority(priorityStr)
	item.IsCompleted = (isCompletedInt == 1)

	parsedStart, err := parseTime(startAtStr)
	if err == nil {
		item.StartAt = parsedStart.Local()
	}

	if endAtStr.Valid && endAtStr.String != "" {
		if parsedEnd, err := parseTime(endAtStr.String); err == nil {
			localEnd := parsedEnd.Local()
			item.EndAt = &localEnd
		}
	}

	if completedAtStr.Valid && completedAtStr.String != "" {
		if parsedComp, err := parseTime(completedAtStr.String); err == nil {
			localComp := parsedComp.Local()
			item.CompletedAt = &localComp
		}
	}

	return &item, nil
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time string: %s", s)
}

func formatNullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
