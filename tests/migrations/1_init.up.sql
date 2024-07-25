CREATE TABLE IF NOT EXISTS users
(
    id        serial PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    pass_hash bytea NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);
