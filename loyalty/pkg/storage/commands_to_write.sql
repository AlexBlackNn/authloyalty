DROP SCHEMA  IF EXISTS loyalty_app CASCADE;

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
CREATE TABLE loyalty_app.loyalty_transactions (
                                                  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                                                  loyalty_uuid uuid REFERENCES loyalty_app.accounts(uuid), -- номер счета (uuid пользователя)
                                                  transaction_amount integer NOT NULL CHECK (transaction_amount > 0), -- сумма операции
                                                  transaction_type operation_type NOT NULL,
                                                  comment text NOT NULL,
                                                  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

----------------------------------------------------------------------
INSERT INTO loyalty_app.accounts (uuid, balance, modified)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 5000.00, '2023-05-28 00:00:00'),
    ('b708f958-a1b4-4e29-b1c8-d8f83b25f0cd', 3000.00, '2023-05-21 00:00:00'),
    ('d0cb3556-a1f7-4b2e-843e-337ab8a4fcbe', 15000.00, '2023-05-22 00:00:00');


INSERT INTO loyalty_app.loyalty_transactions
(loyalty_uuid, transaction_amount, transaction_type, comment, created_at)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 5500.00, 'd', 'registration', '2023-05-30 00:00:00'),
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 500.00, 'w',  'purchase', '2023-05-29 00:00:00'),
    ('b708f958-a1b4-4e29-b1c8-d8f83b25f0cd', 1500.00, 'd','administration', '2023-05-31 00:00:00'),
    ('d0cb3556-a1f7-4b2e-843e-337ab8a4fcbe', 15000.00, 'd', 'purchase','2023-05-30 00:00:00');
------------------------

-- withdraw
UPDATE loyalty_app.accounts
SET balance = balance - 500,
    modified = CURRENT_TIMESTAMP
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

INSERT INTO loyalty_app.loyalty_transactions
(loyalty_uuid, transaction_amount, transaction_type, comment, created_at)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 500.00, 'w', 'purchase', '2023-05-30 00:00:00');

