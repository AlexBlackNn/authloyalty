
CREATE TYPE status AS ENUM ('inProgress', 'successful', 'failed');

CREATE TABLE IF NOT EXISTS users
(
    uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    pass_hash bytea NOT NULL,
    is_admin boolean NOT NULL DEFAULT FALSE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message_status status DEFAULT 'inProgress',
    PRIMARY KEY (uuid, email)
) PARTITION BY HASH (email);


CREATE TABLE users_p1 PARTITION OF users
    FOR VALUES WITH (MODULUS 4, REMAINDER 0);

CREATE TABLE users_p2 PARTITION OF users
    FOR VALUES WITH (MODULUS 4, REMAINDER 1);

CREATE TABLE users_p3 PARTITION OF users
    FOR VALUES WITH (MODULUS 4, REMAINDER 2);

CREATE TABLE users_p4 PARTITION OF users
    FOR VALUES WITH (MODULUS 4, REMAINDER 3);

