
CREATE TYPE status AS ENUM ('inProgress', 'successful', 'failed');

CREATE TABLE IF NOT EXISTS users
(
    uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    pass_hash bytea NOT NULL,
    is_admin boolean NOT NULL DEFAULT FALSE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    full_name text,
    message_status status DEFAULT 'inProgress',
    PRIMARY KEY (email, uuid)
) PARTITION BY HASH (email);


DO $$
    DECLARE
        i INT;
    BEGIN
        FOR i IN 0..3 LOOP
                EXECUTE format('CREATE TABLE users_p%s PARTITION OF users FOR VALUES WITH (MODULUS 4, REMAINDER %s);', i + 1, i);
            END LOOP;
    END $$;
