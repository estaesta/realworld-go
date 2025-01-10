CREATE TABLE following (
    user_id INTEGER NOT NULL,
    follower_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, follower_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (follower_id) REFERENCES user(id) ON DELETE CASCADE
);
