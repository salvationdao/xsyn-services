DROP TYPE IF EXISTS ACCOUNT_TYPE;
CREATE TYPE ACCOUNT_TYPE AS ENUM ('USER', 'SYNDICATE');

CREATE TABLE accounts
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    type       ACCOUNT_TYPE NOT NULL,
    sups       NUMERIC(28)  NOT NULL DEFAULT 0 CHECK (sups >= 0 OR id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' ), -- this is the on chain users / seed account
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);


INSERT INTO accounts (id, sups, type)
SELECT id, sups, 'USER'
FROM users;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_id UUID REFERENCES accounts (id);

UPDATE
    users u
SET account_id = u.id;

ALTER TABLE users
    ALTER COLUMN account_id SET NOT NULL,
    DROP COLUMN sups;

CREATE TABLE syndicates
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    faction_id    UUID        NOT NULL REFERENCES factions (id),
    founded_by_id UUID        NOT NULL REFERENCES users (id),
    name          TEXT        NOT NULL UNIQUE,
    account_id    UUID        NOT NULL REFERENCES accounts (id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS debit_account_id  UUID REFERENCES accounts (id),
    ADD COLUMN IF NOT EXISTS credit_account_id UUID REFERENCES accounts (id);

UPDATE
    transactions t
SET debit_account_id  = t.debit,
    credit_account_id = t.credit;

-- rewrite trigger
CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughfunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF new.debit_account_id = new.credit_account_id THEN
        RAISE EXCEPTION 'unable to transfer to self';
    END IF;

    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT new.debit_account_id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (SELECT accounts.sups >= new.amount
                                                                             FROM accounts
                                                                             WHERE accounts.id = new.debit_account_id)
    INTO enoughfunds;
    -- if enough funds then make the updates to the user table
    IF enoughfunds THEN
        UPDATE accounts SET sups = sups - new.amount WHERE accounts.id = new.debit_account_id;
        UPDATE accounts SET sups = sups + new.amount WHERE accounts.id = new.credit_account_id;
        RETURN new;
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
DROP VIEW IF EXISTS account_ledgers;

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

CREATE INDEX idx_transactions_created_at_descending ON transactions (created_at DESC);
CREATE INDEX idx_transactions_credit ON transactions (credit_account_id);
CREATE INDEX idx_transactions_debit ON transactions (debit_account_id);
CREATE INDEX idx_transactions_debit_credit ON transactions (debit_account_id, credit_account_id);
CREATE INDEX idx_transactions_credit_debit ON transactions (credit_account_id, debit_account_id);
