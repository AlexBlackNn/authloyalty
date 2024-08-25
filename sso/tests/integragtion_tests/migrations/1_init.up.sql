
CREATE TYPE status AS ENUM ('inProgress', 'successful', 'failed');

CREATE TABLE IF NOT EXISTS users
(
    uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    pass_hash bytea NOT NULL,
    is_admin boolean NOT NULL DEFAULT FALSE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    full_name text,
    message_status status DEFAULT 'inProgress'
);