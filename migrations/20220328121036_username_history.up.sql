CREATE TABLE username_history (
    id           UUID        PRIMARY KEY           NOT NULL DEFAULT gen_random_uuid(),
    user_id      UUID        REFERENCES users (id) NOT NULL,
    old_username TEXT                              NOT NULL,
    new_username TEXT                              NOT NULL,
    created_at   TIMESTAMPTZ                       NOT NULL DEFAULT NOW()
);