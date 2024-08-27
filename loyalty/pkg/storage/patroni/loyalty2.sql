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
