-- WARNING!!! SLOW AS SHIT MIGRATION
-- This adds missing indexes to the transactions_old table
-- This should have already been run in the production and staging database

CREATE INDEX IF NOT EXISTS idx_transactions_old_created_at ON transactions_old(created_at timestamptz_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_old_created_desc_at ON transactions_old(created_at timestamptz_ops DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_old_credit ON transactions_old(credit uuid_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_old_debit ON transactions_old(debit uuid_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_old_debit_credit ON transactions_old(debit uuid_ops,credit uuid_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_old_credit_debit ON transactions_old(credit uuid_ops,debit uuid_ops);
