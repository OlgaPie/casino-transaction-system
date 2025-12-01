-- migrations/002_add_transaction_id.sql

ALTER TABLE transactions
ADD COLUMN transaction_id VARCHAR(255) UNIQUE;

CREATE UNIQUE INDEX idx_transactions_transaction_id ON transactions (transaction_id);

ALTER TABLE transactions
ALTER COLUMN transaction_id SET NOT NULL;
