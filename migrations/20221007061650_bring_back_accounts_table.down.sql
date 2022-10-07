DROP INDEX IF EXISTS idx_transactions_created_at_descending;
DROP INDEX IF EXISTS idx_transactions_credit;
DROP INDEX IF EXISTS idx_transactions_debit;
DROP INDEX IF EXISTS idx_transactions_debit_credit;
DROP INDEX IF EXISTS idx_transactions_credit_debit;

ALTER TABLE transactions
    ALTER COLUMN debit_account_id DROP NOT NULL,
    ALTER COLUMN credit_account_id DROP NOT NULL,
    ADD COLUMN debit UUID REFERENCES users(id),
    ADD COLUMN credit UUID REFERENCES users(id);

CREATE INDEX idx_transactions_credit ON transactions (credit);
CREATE INDEX idx_transactions_debit ON transactions (debit);

UPDATE
    transactions t
SET debit = t.debit_account_id,
    credit = t.credit_account_id;

ALTER TABLE transactions
    ALTER COLUMN debit SET NOT NULL,
    ALTER COLUMN credit SET NOT NULL;

ALTER TABLE transactions
    DROP COLUMN debit_account_id,
    DROP COLUMN credit_account_id;

ALTER TABLE users
    ADD COLUMN sups  NUMERIC(28)  NOT NULL DEFAULT 0 CHECK (sups >= 0 OR id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' );

UPDATE users SET sups = (SELECT sups FROM accounts WHERE accounts.id = users.id);

ALTER TABLE users
    DROP COLUMN account_id;

DROP TABLE accounts;
DROP TABLE syndicates;

-- temparary remove account ledgers view
DROP VIEW IF EXISTS account_ledgers;
-- modify account ledger view
CREATE VIEW account_ledgers
            (
             account_id,
             entry_id,
             amount
                )
AS
SELECT transactions.credit,
       transactions.id,
       transactions.amount
FROM transactions
UNION ALL
SELECT transactions.debit,
       transactions.id,
       (0.0 - transactions.amount)
FROM transactions;

DROP TYPE IF EXISTS ACCOUNT_TYPE;
