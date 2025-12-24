
DROP INDEX IF EXISTS idx_transactions_from_account;

DROP INDEX IF EXISTS idx_transactions_to_account;

DROP INDEX IF EXISTS idx_transactions_status;

DROP INDEX IF EXISTS idx_transactions_created_at;

DROP INDEX IF EXISTS idx_transactions_type;

DROP INDEX IF EXISTS idx_transactions_from_status;

ALTER TABLE IF EXISTS transactions DROP CONSTRAINT IF EXISTS fk_from_account;

ALTER TABLE IF EXISTS transactions DROP CONSTRAINT IF EXISTS fk_to_account;

DROP TABLE IF EXISTS transactions;