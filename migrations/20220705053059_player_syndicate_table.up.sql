CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id),
    sups NUMERIC(28) NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

INSERT INTO accounts (user_id, sups)
SELECT id, sups from users;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_id uuid REFERENCES accounts (id);

UPDATE
    users u
SET
    account_id = (SELECT a.id FROM accounts a WHERE a.user_id = u.id);

ALTER TABLE users
    ALTER COLUMN account_id SET NOT NULL;

ALTER TABLE accounts
    DROP COLUMN user_id;

ALTER TABLE users
    DROP COLUMN sups;

CREATE TABLE syndicates(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    founded_by_id uuid not null references users(id),
    account_id uuid not null references accounts(id),
    created_at timestamptz not null default NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS TRANSACTION_TYPE;
CREATE TYPE TRANSACTION_TYPE AS ENUM ('P2P', 'P2S','S2P','S2S'); -- P represent PLAYER, S represent SYNDICATE

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS type TRANSACTION_TYPE NOT NULL DEFAULT 'P2P',
    ADD COLUMN IF NOT EXISTS debit_syndicate uuid REFERENCES syndicates (id),
    ADD COLUMN IF NOT EXISTS credit_syndicate uuid REFERENCES syndicates (id),
    ALTER COLUMN debit DROP NOT NULL,
    ALTER COLUMN credit DROP NOT NULL;

CREATE INDEX idx_transactions_created_at_descending ON transactions(created_at DESC);
CREATE INDEX idx_transactions_debit_credit ON transactions(debit,credit);
CREATE INDEX idx_transactions_debit_syndicate_credit ON transactions(debit_syndicate,credit);
CREATE INDEX idx_transactions_debit_credit_syndicate ON transactions(debit,credit_syndicate);
CREATE INDEX idx_transactions_debit_syndicate_credit_syndicate ON transactions(debit_syndicate,credit_syndicate);

-- rewrite trigger
CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughFunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF NEW.debit = NEW.credit THEN
        RAISE EXCEPTION 'unable to transfer to self';
    END IF;

    -- check debit id and credit id are not missing
    IF NEW.type = 'P2P' THEN
        IF NEW.debit ISNULL OR NEW.credit ISNULL THEN
            RAISE EXCEPTION 'missing debit user id or credit user id';
        END IF;
    ELSEIF NEW.type = 'P2S' THEN
        IF NEW.debit ISNULL OR NEW.credit_syndicate ISNULL THEN
            RAISE EXCEPTION 'missing debit user id or credit syndicate id';
        END IF;
    ELSEIF NEW.type = 'S2P' THEN
        IF NEW.debit_syndicate ISNULL OR NEW.credit ISNULL THEN
            RAISE EXCEPTION 'missing debit syndicate id or credit user id';
        END IF;
    ELSEIF NEW.type = 'S2S' THEN
        IF NEW.debit_syndicate ISNULL OR NEW.credit_syndicate ISNULL THEN
            RAISE EXCEPTION 'missing debit syndicate id or credit syndicate id';
        END IF;
    END IF;

    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT NEW.debit = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (
        CASE
            WHEN NEW.type = 'P2P' OR NEW.TYPE = 'P2S' THEN
                (
                    SELECT accounts.sups >= NEW.amount FROM users
                    INNER JOIN accounts on users.account_id = accounts.id
                    WHERE users.id = NEW.debit
                )
            ELSE
                (
                    SELECT accounts.sups >= NEW.amount
                    FROM syndicates
                             INNER JOIN accounts on syndicates.account_id = accounts.id
                    WHERE syndicates.id = NEW.debit_syndicate
                )
        END
    )
    INTO enoughFunds;
    -- if enough funds then make the updates to the user table
    IF enoughFunds THEN
        IF NEW.type = 'P2P' THEN
            UPDATE accounts SET sups = sups - NEW.amount WHERE accounts.id = (SELECT users.account_id FROM users WHERE users.id = NEW.debit);
            UPDATE accounts SET sups = sups + NEW.amount WHERE accounts.id = (SELECT users.account_id FROM users WHERE users.id = NEW.credit);

        ELSEIF NEW.type = 'P2S' THEN
            UPDATE accounts SET sups = sups - NEW.amount WHERE accounts.id = (SELECT users.account_id FROM users WHERE users.id = NEW.debit);
            UPDATE accounts SET sups = sups + NEW.amount WHERE accounts.id = (SELECT syndicates.account_id FROM syndicates WHERE syndicates.id = NEW.credit_syndicate);

        ELSEIF NEW.type = 'S2P' THEN
            UPDATE accounts SET sups = sups - NEW.amount WHERE accounts.id = (SELECT syndicates.account_id FROM syndicates WHERE syndicates.id = NEW.debit_syndicate);
            UPDATE accounts SET sups = sups + NEW.amount WHERE accounts.id = (SELECT users.account_id FROM users WHERE users.id = NEW.credit);

        ELSEIF NEW.type = 'S2S' THEN
            UPDATE accounts SET sups = sups - NEW.amount WHERE accounts.id = (SELECT syndicates.account_id FROM syndicates WHERE syndicates.id = NEW.debit_syndicate);
            UPDATE accounts SET sups = sups + NEW.amount WHERE accounts.id = (SELECT syndicates.account_id FROM syndicates WHERE syndicates.id = NEW.credit_syndicate);

        END IF;


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