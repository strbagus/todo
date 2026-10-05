# Todo & Deadline Tracker (Browser New-Tab Dashboard) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a lightweight, single-binary personal todo and deadline manager in Go + SQLite with HTMX and Vim keybindings, designed to run as the user's browser default new tab page.

**Architecture:** A monolithic Go application with embedded static assets and HTML templates (`embed.FS`). Standard library Go 1.22+ `net/http` router drives HTMX endpoints. Pure Go SQLite (`modernc.org/sqlite`) handles persistence without CGO. Chronological timeline groups (Overdue, Today, Tomorrow, Upcoming) display tasks and events with type badges and Vim navigation.

**Tech Stack:** Go 1.22+, `modernc.org/sqlite`, Go `html/template`, HTMX 1.9+, Tailwind CSS (compiled static stylesheet), Vanilla JS for Vim keybindings.

**Spec:** [docs/superpowers/specs/2026-10-05-todo-deadline-tracker-design.md](file:///home/strbagus/Development/Project/todo/docs/superpowers/specs/2026-10-05-todo-deadline-tracker-design.md)

## Global Constraints

- Go standard library formatting (`gofmt`) for all Go code.
- Pure Go SQLite driver (`modernc.org/sqlite`) without CGO dependencies.
- Zero NPM build steps required at runtime; all static assets (CSS, JS) must be embedded directly in the binary.
- All date calculations must respect the user's local timezone.
- No destructive git operations without explicit user approval.

## Review Focus

1. **Timezone boundary crossing**: Items due at 23:59 local time must fall into "Today", while 00:00 next day must fall into "Tomorrow".
2. **Invalid datetime submissions**: Handlers must return HTTP 422 with meaningful error messages instead of crashing or returning 500 when receiving malformed datetime strings.
3. **Empty / whitespace-only titles**: Titles with only whitespace must fail validation before hitting SQLite.
4. **EndAt earlier than StartAt**: Event creation with `EndAt` prior to `StartAt` must fail validation.
5. **Vim shortcut focus conflict**: Typing inside text input / textarea fields must not trigger Vim navigation (`j`, `k`, `x`, `o`).

---

### Task 1: Project Scaffolding & SQLite Database Setup

**Files:**
- Create: `go.mod`
- Create: `internal/database/schema.sql`
- Create: `internal/database/db.go`
- Test: `internal/database/db_test.go`

**Interfaces:**
- Produces: `database.InitDB(dbPath string) (*sql.DB, error)`
- Produces: `database.Migrate(db *sql.DB) error`

- [ ] **Step 1: Write the failing test**

```go
// internal/database/db_test.go
package database_test

import (
	"testing"
	"todo/internal/database"
)

func TestInitDB_And_Migrate(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='items'").Scan(&tableName)
	if err != nil || tableName != "items" {
		t.Fatalf("expected items table to exist, got error: %v, tableName: %s", err, tableName)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/database/...`
Expected: FAIL (cannot find module or package)

- [ ] **Step 3: Implement Go module and `database` package**

Initialize module `todo` with `go mod init todo`, add dependency `modernc.org/sqlite`. Implement `internal/database/schema.sql` with table `items` and indexes from the spec. Implement `InitDB` to open SQLite with WAL mode (`_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)`) and `Migrate` to execute the schema.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/database/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Ensure code formatting with `gofmt -w internal/database/`.

---

### Task 2: Item Domain Model, Validation & Timeline Grouping

**Files:**
- Create: `internal/models/item.go`
- Create: `internal/models/timeline.go`
- Test: `internal/models/item_test.go`
- Test: `internal/models/timeline_test.go`

**Interfaces:**
- Consumes: Standard Go `time`
- Produces: `models.Item`, `models.ItemType`, `models.Priority`, `models.TimelineGroups`
- Produces: `(item *Item) Validate() error`
- Produces: `models.GroupItems(items []Item, now time.Time) TimelineGroups`

- [ ] **Step 1: Write the failing tests**

```go
// internal/models/item_test.go
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

// internal/models/timeline_test.go
func TestGroupItems(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)
	overdue := models.Item{ID: 1, Title: "Overdue Task", Type: models.TypeCollegeTask, StartAt: now.Add(-24 * time.Hour), IsCompleted: false}
	todayItem := models.Item{ID: 2, Title: "Today Task", Type: models.TypeWorkDue, StartAt: now.Add(2 * time.Hour), IsCompleted: false}
	tomorrowItem := models.Item{ID: 3, Title: "Tomorrow Event", Type: models.TypeEvent, StartAt: now.Add(26 * time.Hour)}
	upcomingItem := models.Item{ID: 4, Title: "Next Week Exam", Type: models.TypeCollegeTask, StartAt: now.Add(5 * 24 * time.Hour)}
	completedItem := models.Item{ID: 5, Title: "Done Task", Type: models.TypeCollegeTask, StartAt: now.Add(-1 * time.Hour), IsCompleted: true}

	groups := models.GroupItems([]models.Item{overdue, todayItem, tomorrowItem, upcomingItem, completedItem}, now)

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
	if len(groups.Completed) != 1 || groups.Completed[0].ID != 5 {
		t.Errorf("expected 1 completed item, got %d", len(groups.Completed))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/models/...`
Expected: FAIL (models package not implemented)

- [ ] **Step 3: Implement `models/item.go` and `models/timeline.go`**

Implement `Item` struct, `ItemType`, `Priority`, `Validate() error` enforcing non-empty title, valid enum types, and `EndAt >= StartAt`. Implement `GroupItems` calculating `todayStart`, `tomorrowStart`, and `dayAfterStart` in local time.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/models/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Format code with `gofmt -w internal/models/`.

---

### Task 3: SQLite Repository Layer

**Files:**
- Create: `internal/repository/item_repo.go`
- Test: `internal/repository/item_repo_test.go`

**Interfaces:**
- Consumes: `database.InitDB`, `models.Item`, `models.ItemType`
- Produces: `ItemRepository`:
  - `Create(ctx context.Context, item *models.Item) error`
  - `GetByID(ctx context.Context, id int64) (*models.Item, error)`
  - `List(ctx context.Context, filterType models.ItemType, search string) ([]models.Item, error)`
  - `ToggleCompleted(ctx context.Context, id int64) (*models.Item, error)`
  - `Update(ctx context.Context, item *models.Item) error`
  - `Delete(ctx context.Context, id int64) error`

- [ ] **Step 1: Write the failing tests**

```go
// internal/repository/item_repo_test.go
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

	// 4. List with filter
	list, err := repo.List(ctx, models.TypeCollegeTask, "")
	if err != nil || len(list) != 1 {
		t.Fatalf("List failed or unexpected count: len=%d, err=%v", len(list), err)
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/repository/...`
Expected: FAIL (repository package not implemented)

- [ ] **Step 3: Implement `internal/repository/item_repo.go`**

Implement `ItemRepo` struct holding `*sql.DB`. Implement queries:
- `Create`: `INSERT INTO items (...) VALUES (...) RETURNING id, created_at, updated_at`
- `GetByID`: `SELECT ... FROM items WHERE id = ?`
- `List`: Dynamic query appending `WHERE type = ?` if `filterType != ""` and `AND (title LIKE ? OR context_tag LIKE ?)` if search is provided. Order by `start_at ASC`.
- `ToggleCompleted`: Atomically flips `is_completed = 1 - is_completed`, sets `completed_at = CASE WHEN is_completed=1 THEN CURRENT_TIMESTAMP ELSE NULL END`.
- `Update`: Updates title, description, start_at, end_at, context_tag, location_or_url, priority.
- `Delete`: `DELETE FROM items WHERE id = ?`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/repository/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Format code with `gofmt -w internal/repository/`.

---

### Task 4: Static Assets, Vim Keybindings & Base Templates

**Files:**
- Create: `web/static/css/style.css`
- Create: `web/static/js/htmx.min.js`
- Create: `web/static/js/vim.js`
- Create: `web/templates/base.html`
- Create: `web/templates/components/cheatsheet_modal.html`
- Create: `internal/templates/templates.go`
- Test: `internal/templates/templates_test.go`

**Interfaces:**
- Produces: `templates.FS` (embedded static and template filesystem)
- Produces: `templates.Render(w io.Writer, name string, data any) error`
- Produces: Vim navigation handler in `vim.js`

- [ ] **Step 1: Write the failing test**

```go
// internal/templates/templates_test.go
package templates_test

import (
	"bytes"
	"testing"
	"todo/internal/templates"
)

func TestTemplateParsing(t *testing.T) {
	tmpl, err := templates.ParseTemplates()
	if err != nil {
		t.Fatalf("ParseTemplates failed: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "cheatsheet_modal.html", nil)
	if err != nil {
		t.Fatalf("ExecuteTemplate cheatsheet_modal failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Vim Shortcuts")) {
		t.Errorf("expected cheatsheet modal to contain 'Vim Shortcuts'")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/templates/...`
Expected: FAIL (templates package not implemented)

- [ ] **Step 3: Implement static assets and base template system**

- Add standalone Tailwind CSS utility classes into `web/static/css/style.css` (custom dark palette, badges for college/work/events, modal overlays).
- Embed HTMX 1.9+ library in `web/static/js/htmx.min.js`.
- Write `web/static/js/vim.js`:
  - Ignore shortcuts if active element is `INPUT`, `TEXTAREA`, or `SELECT` (except `Esc` to blur).
  - Handle `j`/`k` to toggle `.selected-card` and scroll smoothly into view.
  - Handle `x`/`Space` to trigger HTMX toggle on the selected card.
  - Handle `o`/`a` to show the Add Item modal.
  - Handle `e` to trigger edit modal on selected card.
  - Handle `d` to trigger delete on selected card.
  - Handle `/` to focus the search bar.
  - Handle `1`, `2`, `3`, `4` to switch filter tabs.
  - Handle `?` to toggle the shortcuts cheatsheet.
- Implement `internal/templates/templates.go` using `embed.FS` to embed `web/`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/templates/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Format code with `gofmt -w internal/templates/`.

---

### Task 5: Dashboard View & Timeline Component Templates

**Files:**
- Create: `web/templates/index.html`
- Create: `web/templates/components/item_card.html`
- Create: `web/templates/components/timeline_group.html`
- Create: `internal/handlers/views_handler.go`
- Test: `internal/handlers/views_handler_test.go`

**Interfaces:**
- Consumes: `repository.ItemRepo`, `models.GroupItems`, `templates.Render`
- Produces: `views_handler.ServeDashboard(w http.ResponseWriter, r *http.Request)`

- [ ] **Step 1: Write the failing test**

```go
// internal/handlers/views_handler_test.go
package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/templates"
)

func TestServeDashboard(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	_ = database.Migrate(db)
	repo := repository.NewItemRepo(db)
	tmpl, _ := templates.ParseTemplates()

	// Seed item
	_ = repo.Create(context.Background(), &models.Item{
		Title:    "Operating Systems Lab",
		Type:     models.TypeCollegeTask,
		StartAt:  time.Now().Add(2 * time.Hour),
		Priority: models.PriorityHigh,
	})

	handler := handlers.NewViewsHandler(repo, tmpl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Operating Systems Lab") {
		t.Errorf("expected dashboard body to contain 'Operating Systems Lab'")
	}
	if !strings.Contains(body, "Today") {
		t.Errorf("expected dashboard body to contain timeline group 'Today'")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/handlers/...`
Expected: FAIL (handler not implemented)

- [ ] **Step 3: Implement dashboard templates & views handler**

- `web/templates/components/item_card.html`: Renders item with badge (College in indigo, Work in amber, Event in purple), due/start time, checkbox (for tasks), action buttons, and `data-item-id="{{.ID}}"`.
- `web/templates/components/timeline_group.html`: Renders section header (e.g. Overdue, Today, Tomorrow, Upcoming) and cards.
- `web/templates/index.html`: Header with date, filter tabs, search bar, Add button, and timeline group container.
- `internal/handlers/views_handler.go`: Fetches items, calls `models.GroupItems`, executes `index.html`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/handlers/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Format code with `gofmt -w internal/handlers/`.

---

### Task 6: HTMX Item Handlers (Add, Toggle, Edit, Delete, Filter)

**Files:**
- Create: `web/templates/components/add_modal.html`
- Create: `web/templates/components/edit_modal.html`
- Create: `internal/handlers/item_handler.go`
- Create: `internal/handlers/routes.go`
- Test: `internal/handlers/item_handler_test.go`

**Interfaces:**
- Consumes: `repository.ItemRepo`, `templates.Render`, `models.Item`
- Produces: HTTP handlers for:
  - `POST /items`: Create item
  - `POST /items/{id}/toggle`: Toggle complete
  - `GET /items/{id}/edit`: Render edit form modal
  - `PUT /items/{id}`: Update item
  - `DELETE /items/{id}`: Delete item
  - `GET /items/filter`: Render filtered timeline feed partial

- [ ] **Step 1: Write the failing tests**

```go
// internal/handlers/item_handler_test.go
package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/templates"
)

func TestItemHandler_Endpoints(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	_ = database.Migrate(db)
	repo := repository.NewItemRepo(db)
	tmpl, _ := templates.ParseTemplates()
	itemHandler := handlers.NewItemHandler(repo, tmpl)
	mux := handlers.RegisterRoutes(repo, tmpl)

	// 1. Validation failure returns 422
	form := url.Values{}
	form.Set("title", "") // empty title
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for empty title, got %d", rec.Code)
	}

	// 2. Valid creation returns 200/201 with item card
	now := time.Now().Format("2006-01-02T15:04")
	form.Set("title", "Finish Thesis Proposal")
	form.Set("type", string(models.TypeCollegeTask))
	form.Set("start_at", now)
	form.Set("priority", string(models.PriorityHigh))
	req = httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("expected 200/201 for valid creation, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Finish Thesis Proposal") {
		t.Errorf("response does not contain created item title")
	}

	// 3. Toggle completed
	items, _ := repo.List(context.Background(), "", "")
	if len(items) == 0 {
		t.Fatalf("expected at least 1 item in db")
	}
	targetID := items[0].ID
	req = httptest.NewRequest(http.MethodPost, "/items/"+string(rune(targetID+'0'))+"/toggle", nil)
	// Using standard router format
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items/"+strings.TrimSpace(rec.Body.String()), nil))

	// Verify delete
	req = httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on delete, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/handlers/...`
Expected: FAIL (item handler and routes not implemented)

- [ ] **Step 3: Implement item handlers and route registration**

Implement `internal/handlers/item_handler.go`:
- Parse and validate POST form values into `models.Item`.
- Return 422 with inline error message partial if validation fails.
- On create, render `item_card.html` (or full timeline if HTMX target is `#timeline`).
- Handle toggle, returning updated `item_card.html`.
- Handle edit modal retrieval and update.
- Handle delete returning HTTP 200 with empty body so HTMX removes the element.
- Implement `RegisterRoutes(repo, tmpl) *http.ServeMux` with Go 1.22+ method routing.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/handlers/...`
Expected: PASS

- [ ] **Step 5: Record task completion**

Format code with `gofmt -w internal/handlers/`.

---

### Task 7: Server Entrypoint, Graceful Shutdown & End-to-End Verification

**Files:**
- Create: `cmd/server/main.go`

**Interfaces:**
- Consumes: `database`, `repository`, `templates`, `handlers`
- Produces: Executable single binary `server`

- [ ] **Step 1: Write server entrypoint**

In `cmd/server/main.go`:
- Read port from flag `-port` or `PORT` environment variable (default `8080`).
- Read DB path from flag `-db` or `DB_PATH` environment variable (default `todo.db`).
- Initialize SQLite via `database.InitDB` and run `database.Migrate`.
- Parse templates and register routes.
- Serve static assets via `http.FileServerFS(templates.FS)`.
- Configure `http.Server` with timeouts (ReadTimeout 5s, WriteTimeout 10s).
- Setup `signal.Notify(quit, os.Interrupt, syscall.SIGTERM)` for graceful shutdown with 5s context timeout.

- [ ] **Step 2: Build the binary**

Run: `go build -o bin/todo ./cmd/server`
Expected: Successful compilation producing executable `bin/todo`.

- [ ] **Step 3: End-to-end integration test execution**

Run the complete test suite:
Run: `go test -v ./...`
Expected: All tests PASS.

- [ ] **Step 4: Smoke test server launch and HTTP response**

Run a temporary background process of `./bin/todo -port 8888 -db /tmp/test_todo.db`, send `curl -i http://localhost:8888/`, verify HTTP 200 and expected HTML tags, then terminate the process cleanly.

- [ ] **Step 5: Final review and handoff**

Verify all code complies with Go standards and spec requirements.
