CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cid CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    alias TEXT NOT NULL,
    original_url TEXT NOT NULL UNIQUE,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    UNIQUE (user_id, original_url)
);
