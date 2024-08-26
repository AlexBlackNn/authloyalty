CREATE SCHEMA IF NOT EXISTS loyalty_app;

CREATE TABLE IF NOT EXISTS loyalty_app.loyalty
(
    uuid uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    value integer
);