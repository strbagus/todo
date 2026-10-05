# Design Specification: Todo & Deadline Tracker (Browser New-Tab Dashboard)

**Date**: 2026-10-05  
**Status**: Draft (Approved in Brainstorming)  
**Target Environment**: Linux, Go 1.22+, SQLite, Modern Web Browser  

---

## 1. Overview & Goals

The goal of this project is to build a fast, single-binary, lightweight personal task and event tracker tailored to run as the user's **default browser new-tab page**. It provides an instant, unified, chronological overview of upcoming deadlines and events across college, work, and personal schedules.

### Primary Objectives
1. **Unified Timeline Feed**: Clearly visualize upcoming commitments grouped chronologically into Overdue, Today, Tomorrow, and Upcoming.
2. **Type Categorization**: Distinctly manage three core item types:
   - **College Tasks**: Assignments, projects, and exams with due dates, course codes, and completion status.
   - **Work Dues**: Work deliverables, sprint items, and tasks with deadlines, project tags, priority levels, and completion status.
   - **Events**: Scheduled occurrences with date, start time, optional end time/duration, and location or meeting links (not completion-checked).
3. **Frictionless Entry & Keyboard-First UX**: Full Vim-style keybindings (`j`/`k`, `x`, `o`, `d`, `/`, etc.) for seamless interaction without touching the mouse.
4. **Zero-Setup Single Binary**: Built with Go, embedding SQLite (`modernc.org/sqlite` pure Go), Go HTML templates, HTMX, and Tailwind CSS into a single executable with zero external runtime dependencies.

---

## 2. Architecture & Tech Stack

```
todo/
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint & HTTP server
├── internal/
│   ├── database/
│   │   ├── db.go            # SQLite connection setup & migration runner
│   │   └── schema.sql       # Schema definition
│   ├── models/
│   │   └── item.go          # Item entity, validation, and group classification
│   ├── repository/
│   │   ├── item_repo.go     # SQLite CRUD queries & filtering
│   │   └── item_repo_test.go# In-memory SQLite integration tests
│   └── handlers/
│       ├── item_handler.go  # HTMX partial and action handlers
│       ├── views_handler.go # Main dashboard page renderer
│       └── routes.go        # HTTP route registration (Go 1.22+ mux)
├── web/
│   ├── static/
│   │   ├── css/style.css    # Tailwind CSS bundle
│   │   └── js/htmx.min.js   # Embedded HTMX library
│   └── templates/
│       ├── base.html        # HTML layout shell with Vim shortcuts & theme
│       ├── index.html       # Timeline page template
│       └── components/      # HTMX reusable partials
│           ├── item_card.html
│           ├── add_modal.html
│           ├── edit_modal.html
│           ├── cheatsheet_modal.html
│           └── timeline_group.html
├── docs/
│   └── superpowers/specs/2026-10-05-todo-deadline-tracker-design.md
├── go.mod
└── go.sum
```

### Technology Selections & Rationale
- **Backend Language**: Go 1.22+ utilizing enhanced standard library routing (`net/http.ServeMux` with method + path patterns like `POST /items/{id}/toggle`).
- **Database**: SQLite via `modernc.org/sqlite` (Pure Go, 100% CGO-free, enables effortless compilation across systems without GCC).
- **Frontend Template Engine**: Go standard `html/template` with embedded files via `embed.FS`.
- **Frontend Interactivity**: HTMX 1.9+ for dynamic updates (add, toggle, filter, delete) without full page reloads.
- **Styling**: Tailwind CSS (bundled statically into `web/static/css/style.css`).
- **Portability**: The compiled binary embeds all templates, CSS, and JS, persisting data to a local `todo.db` SQLite file.

---

## 3. Data Model & Database Schema

### Database Table: `items`

```sql
CREATE TABLE IF NOT EXISTS items (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    type            TEXT NOT NULL CHECK(type IN ('college_task', 'work_due', 'event')),
    start_at        DATETIME NOT NULL,              -- Due datetime for tasks; Start datetime for events
    end_at          DATETIME,                       -- Optional end datetime (used for events)
    context_tag     TEXT NOT NULL DEFAULT '',       -- Course code (e.g. 'CS101') or project ('Client-X')
    location_or_url TEXT NOT NULL DEFAULT '',       -- Submission link, Zoom URL, or room location
    priority        TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high')),
    is_completed    INTEGER NOT NULL DEFAULT 0,     -- 0 (false) or 1 (true); only applies to dues/tasks
    completed_at    DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_start_at ON items(start_at);
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_completed ON items(is_completed);
```

### Go Entity Definition

```go
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
```

---

## 4. Timeline Grouping & Business Logic

The timeline view groups uncompleted items based on `StartAt` relative to the current local date:

1. **Overdue**:
   - Condition: `IsCompleted == false AND Type IN ('college_task', 'work_due') AND StartAt < TodayStart`
   - Visual: Alert red banner/container at top of page with count indicator.
2. **Today**:
   - Condition: `StartAt >= TodayStart AND StartAt < TomorrowStart` (excluding completed tasks).
3. **Tomorrow**:
   - Condition: `StartAt >= TomorrowStart AND StartAt < DayAfterTomorrowStart`.
4. **Upcoming**:
   - Condition: `StartAt >= DayAfterTomorrowStart` (sorted chronologically).
5. **Completed** *(Collapsible)*:
   - Condition: `IsCompleted == true`.
   - Rendered at the bottom with strike-through styling to keep the active viewport clean.

---

## 5. UI/UX & Interactions

### Layout Structure
- **Top Header**:
  - Current Date & Greeting (e.g. *"Monday, October 5"*).
  - Filter Tabs: `All`, `College`, `Work`, `Events`.
  - Search Input (`/` shortcut).
  - `+ Add New (o)` button & Keyboard Shortcuts `(?)` button.
- **Main Feed**:
  - Rendered timeline groups with sticky date headers.
  - Item cards featuring distinct badge palettes:
    - **College Task**: Indigo badge (`[College: CS101]`), due time, submission link button, completion checkbox.
    - **Work Due**: Amber badge (`[Work: Feature-A]`), priority tag, due time, completion checkbox.
    - **Event**: Purple badge (`[Event]`), scheduled time span (`14:00 - 15:30`), location/URL button, calendar icon (no checkbox).

### Vim Keyboard Shortcuts Specification
Implemented via lightweight vanilla JS event listener on the window:
- `j`: Select next item in the timeline.
- `k`: Select previous item in the timeline.
- `gg`: Jump selection to the top item.
- `G`: Jump selection to the bottom item.
- `x` or `Space`: Toggle complete status for the selected task/due (HTMX trigger).
- `o` or `a`: Open the "Add New Item" modal.
- `e`: Open the edit modal for the currently selected item.
- `d` or `dd`: Delete selected item (with confirmation prompt/toast).
- `/`: Focus search/filter input.
- `1` / `2` / `3` / `4`: Switch filter tabs (`1: All`, `2: College`, `3: Work`, `4: Events`).
- `Esc`: Close modal, blur inputs, or clear selection.
- `?`: Toggle keyboard shortcuts cheatsheet modal.

### HTMX Endpoints
- `GET /`: Full dashboard render.
- `GET /items?filter={college|work|event}&q={search}`: Returns filtered timeline partial.
- `POST /items`: Creates item, returns updated timeline group or prepended item partial.
- `POST /items/{id}/toggle`: Toggles completion status, swaps item card or shifts to Completed section.
- `GET /items/{id}/edit`: Returns item edit form modal partial.
- `PUT /items/{id}`: Updates item, swaps item card in place.
- `DELETE /items/{id}`: Deletes item, fades out and removes card from DOM.

---

## 6. Error Handling & Validation

1. **Input Validation**:
   - `Title`: Required, trimmed length between 1 and 255 characters.
   - `Type`: Must be one of `college_task`, `work_due`, `event`.
   - `StartAt`: Required valid RFC3339 datetime.
   - `EndAt`: Optional, but if present must be `>= StartAt`.
   - `Priority`: Must be `low`, `medium`, or `high`.
2. **HTMX Validation Feedback**:
   - If validation fails on `POST /items`, HTTP status `422 Unprocessable Entity` is returned along with the form partial containing inline field error highlights.
3. **Database Concurrency & Graceful Shutdown**:
   - Use SQLite WAL (Write-Ahead Logging) mode: `PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;` for optimal local read/write performance.
   - Server listens for `SIGINT` / `SIGTERM` and executes graceful HTTP server shutdown and SQLite connection closure.

---

## 7. Testing Strategy

1. **Unit Tests**:
   - Validation rules in `internal/models/item_test.go`.
   - Timeline grouping logic in `internal/models/timeline_test.go` with fixed mocked dates.
2. **Integration Tests**:
   - SQLite queries in `internal/repository/item_repo_test.go` executed against in-memory SQLite (`file::memory:?cache=shared`).
3. **Handler Tests**:
   - HTTP route and template rendering tests using Go's `net/http/httptest` package in `internal/handlers/item_handler_test.go`.
