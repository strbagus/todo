package repository_test

import (
	"context"
	"testing"
	"time"
	"todo/internal/database"
	"todo/internal/models"
	"todo/internal/repository"
)

func setupTestDB(t *testing.T) *repository.ItemRepo {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return repository.NewItemRepo(db)
}

func TestItemRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := setupTestDB(t)

	now := time.Now().Truncate(time.Second)
	item := &models.Item{
		Title:       "CS101 Homework",
		Description: "Chapters 1-3",
		Type:        models.TypeCollegeTask,
		StartAt:     now,
		ContextTag:  "CS101",
		Priority:    models.PriorityHigh,
	}

	// 1. Create
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected item.ID to be assigned")
	}

	// 2. GetByID
	fetched, err := repo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Title != item.Title || fetched.IsCompleted {
		t.Fatalf("unexpected fetched item: %+v", fetched)
	}

	// 3. ToggleCompleted
	toggled, err := repo.ToggleCompleted(ctx, item.ID)
	if err != nil {
		t.Fatalf("ToggleCompleted failed: %v", err)
	}
	if !toggled.IsCompleted || toggled.CompletedAt == nil {
		t.Fatalf("expected item to be completed, got %+v", toggled)
	}

	// 3b. Toggle again to verify uncompleting
	unToggled, err := repo.ToggleCompleted(ctx, item.ID)
	if err != nil {
		t.Fatalf("ToggleCompleted (2nd) failed: %v", err)
	}
	if unToggled.IsCompleted || unToggled.CompletedAt != nil {
		t.Fatalf("expected item to be uncompleted, got %+v", unToggled)
	}

	// 4. List with filter and search
	list, err := repo.List(ctx, models.TypeCollegeTask, "Homework")
	if err != nil || len(list) != 1 {
		t.Fatalf("List with search failed or unexpected count: len=%d, err=%v", len(list), err)
	}
	listEmpty, err := repo.List(ctx, models.TypeWorkDue, "")
	if err != nil || len(listEmpty) != 0 {
		t.Fatalf("expected 0 work dues, got len=%d", len(listEmpty))
	}

	// 5. Delete
	if err := repo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.GetByID(ctx, item.ID)
	if err == nil {
		t.Fatalf("expected error when getting deleted item")
	}
}
