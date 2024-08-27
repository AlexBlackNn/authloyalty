-- DROP SCHEMA  IF EXISTS loyalty_app CASCADE;

CREATE SCHEMA IF NOT EXISTS loyalty_app;

CREATE TYPE operation_type AS ENUM ('w', 'd'); --withdrew/ deposit

CREATE TABLE IF NOT EXISTS loyalty_app.accounts
(
uuid uuid PRIMARY KEY,  -- номер счета (uuid пользователя)
modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
balance DECIMAL(12,0) CONSTRAINT not_negative_loyalty_value CHECK (balance >= 0) -- — текущий остаток на счёте
);

-- хранит записи о всех транзакциях по лояльности.
-- после успешного пополения, уменьшения балов нужно добавить запись о транзакции
-- в таблицу loyalty_transactions. Баланс не должен быть отрицательным,
-- но может быть равным нулю, а loyalty_amount должны быть положительными.
CREATE TABLE IF NOT EXISTS loyalty_app.loyalty_transactions (
id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
loyalty_uuid uuid REFERENCES loyalty_app.accounts(uuid), -- номер счета (uuid пользователя)
transaction_amount integer NOT NULL CHECK (transaction_amount > 0), -- сумма операции
transaction_type operation_type NOT NULL,
comment text NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
