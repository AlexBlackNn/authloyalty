-- withdraw 500 
-- открываем транзакцию 
BEGIN;

-- Получить данные о существовании записи пользователя с uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7'
SELECT balance
FROM loyalty_app.accounts
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Если записи нет, то выдаем ошибку. Так как нельзя вычисть скидку с счета, которого пока нет.  
ROLLBACK;
-- Если запись есть, и balance - 500 >= 0 то делаем 
UPDATE loyalty_app.accounts
SET balance =  balance - 5000,
    modified = CURRENT_TIMESTAMP
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Делаем пометку в историю 
INSERT INTO loyalty_app.loyalty_transactions
(loyalty_uuid, transaction_amount, transaction_type, comment, created_at)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 5000.00, 'w', 'purchase', '2023-05-30 00:00:00');
-- закрываем транзакцию 
COMMIT;
---
-- withdraw 500
-- открываем транзакцию
-- Если СУБД обнаружит потенциальный конфликт при попытке завершения SERIALIZABLE-транзакции, она просто отменит
-- транзакцию, а это может привести к повышенному количеству откатов транзакций. Поэтому, используя SERIALIZABLE,
-- пришлось бы добавить обработку откатов транзакций в случае ошибок сериализации и повторный запуск этих транзакций.
-- Это усложняет код и снижает эффективность работы приложения из-за постоянных ошибок и перезапусков транзакций.
-- BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
-- Получить данные о существовании записи пользователя с uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7'
SELECT balance
FROM loyalty_app.accounts
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Если записи нет, то выдаем ошибку. Так как нельзя вычисть скидку с счета, которого пока нет.
ROLLBACK;
-- Если запись есть, и balance - 500 >= 0 то делаем
UPDATE loyalty_app.accounts
SET balance =  balance-2000,						-- 5000 - 1000 result will be 4000 after concurrent operation
    modified = CURRENT_TIMESTAMP
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Делаем пометку в историю
INSERT INTO loyalty_app.loyalty_transactions
(loyalty_uuid, transaction_amount, transaction_type, comment, created_at)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 2000.00, 'w', 'purchase', '2023-05-30 00:00:00');
-- закрываем транзакцию
COMMIT;

-----
-- withdraw 500
-- открываем транзакцию
-- BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
-- Получить данные о существовании записи пользователя с uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7'

-- SELECT FOR SHARE  (кажется блокировка тут не нужна, так как постгрес сделает ее автоматически.)

-- Если записи нет, то выдаем ошибку. Так как нельзя вычисть скидку с счета, которого пока нет.
-- Есть транзакция, которая должна обновить запись, если она существует, а если нет — создать новую.
-- Одновременно может быть запущено несколько таких транзакций. Если две такие операции выполняются одновременно,
-- они могут обе решить, что запись не существует, и попытаться создать две новые записи. Необходимо так настроить
-- работу транзакции, чтобы обработка возможных сбоев и повторный запуск не потребовались.
SELECT *
FROM loyalty_app.accounts
WHERE uuid = '19d31ff1-e0f0-4eda-b52c-a68e2ab3eae7'
    FOR UPDATE;
---
SELECT *
FROM loyalty_app.accounts
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7'
    FOR UPDATE;

SELECT balance
FROM loyalty_app.accounts
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Если записи нет, то выдаем ошибку. Так как нельзя вычисть скидку с счета, которого пока нет.
ROLLBACK;
-- Если запись есть, и balance - 500 >= 0 то делаем
UPDATE loyalty_app.accounts
SET balance =  balance-2000,						-- 5000 - 1000 result will be 4000 after concurrent operation
    modified = CURRENT_TIMESTAMP
WHERE uuid = 'e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7';

-- Делаем пометку в историю
INSERT INTO loyalty_app.loyalty_transactions
(loyalty_uuid, transaction_amount, transaction_type, comment, created_at)
VALUES
    ('e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7', 2000.00, 'w', 'purchase', '2023-05-30 00:00:00');
-- закрываем транзакцию
COMMIT;

