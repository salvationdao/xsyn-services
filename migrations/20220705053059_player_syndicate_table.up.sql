DROP TYPE IF EXISTS ACCOUNT_TYPE;
CREATE TYPE ACCOUNT_TYPE AS ENUM ('USER', 'SYNDICATE');

CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type ACCOUNT_TYPE NOT NULL,
    sups NUMERIC(28) NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

INSERT INTO accounts (id, sups, type)
SELECT id, sups, 'USER' from users;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_id uuid REFERENCES accounts (id);

UPDATE
    users u
SET
    account_id = u.id;

ALTER TABLE users
    ALTER COLUMN account_id SET NOT NULL,
    DROP COLUMN sups;

CREATE TABLE syndicates(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id uuid not null references factions (id),
    founded_by_id uuid not null references users(id),
    name text NOT NULL UNIQUE,
    account_id uuid not null references accounts(id),
    created_at timestamptz not null default NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS debit_account_id uuid REFERENCES accounts (id),
    ADD COLUMN IF NOT EXISTS credit_account_id uuid REFERENCES accounts (id);

UPDATE
    transactions t
SET
    debit_account_id = t.debit,
    credit_account_id = t.credit;


-- rewrite trigger
CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughFunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF NEW.debit_account_id = NEW.credit_account_id THEN
        RAISE EXCEPTION 'unable to transfer to self';
    END IF;

    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT NEW.debit_account_id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (
        SELECT accounts.sups >= NEW.amount
        FROM accounts WHERE accounts.id = NEW.debit_account_id
    )
    INTO enoughFunds;
    -- if enough funds then make the updates to the user table
    IF enoughFunds THEN
        UPDATE accounts SET sups = sups - NEW.amount WHERE accounts.id = NEW.debit_account_id;
        UPDATE accounts SET sups = sups + NEW.amount WHERE accounts.id = NEW.credit_account_id;
        RETURN NEW;
    ELSE
        RAISE EXCEPTION 'not enough funds';
    END IF;
    -- if not enough funds,
END
$check_balances$
    LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trigger_check_balance
    ON transactions;
CREATE TRIGGER trigger_check_balance
    BEFORE INSERT
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE check_balances();

-- temparary remove account ledgers view
DROP VIEW account_ledgers;

-- modify account ledger view
CREATE VIEW account_ledgers
            (
             account_id,
             entry_id,
             amount
                )
AS
SELECT transactions.credit_account_id,
       transactions.id,
       transactions.amount
FROM transactions
UNION ALL
SELECT transactions.debit_account_id,
       transactions.id,
       (0.0 - transactions.amount)
FROM transactions;

DROP INDEX transactions_credit_idx;
DROP INDEX transactions_debit_idx;

ALTER TABLE transactions
    ALTER COLUMN debit_account_id SET NOT NULL,
    ALTER COLUMN credit_account_id SET NOT NULL,
    DROP COLUMN IF EXISTS debit,
    DROP COLUMN IF EXISTS credit;

CREATE INDEX idx_transactions_created_at_descending ON transactions(created_at DESC);
CREATE INDEX idx_transactions_credit ON transactions(credit_account_id);
CREATE INDEX idx_transactions_debit ON transactions(debit_account_id);
CREATE INDEX idx_transactions_debit_credit ON transactions(debit_account_id,credit_account_id);
CREATE INDEX idx_transactions_credit_debit ON transactions(credit_account_id,debit_account_id);