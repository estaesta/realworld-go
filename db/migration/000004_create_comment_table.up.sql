CREATE TABLE comment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    author_id INTEGER NOT NULL,
    article_id INTEGER NOT NULL,
    FOREIGN KEY (author_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES article(id) ON DELETE CASCADE
);
