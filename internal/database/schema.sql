CREATE TABLE IF NOT EXISTS items (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    type            TEXT NOT NULL CHECK(type IN ('college_task', 'work_due', 'event')),
    start_at        DATETIME NOT NULL,
    end_at          DATETIME,
    context_tag     TEXT NOT NULL DEFAULT '',
    location_or_url TEXT NOT NULL DEFAULT '',
    priority        TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high')),
    is_completed    INTEGER NOT NULL DEFAULT 0,
    completed_at    DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_start_at ON items(start_at);
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_completed ON items(is_completed);
