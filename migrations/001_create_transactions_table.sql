-- migrations/001_create_transactions_table.sql
CREATE TABLE transactions
(
    id               SERIAL PRIMARY KEY,
    user_id          VARCHAR(255)   NOT NULL,
    transaction_type VARCHAR(10)    NOT NULL CHECK (transaction_type IN ('bet', 'win')),
    amount           NUMERIC(15, 2) NOT NULL,
    "timestamp"      TIMESTAMPTZ    NOT NULL
);

CREATE INDEX idx_transactions_user_id ON transactions (user_id);