DROP TYPE IF EXISTS ACCOUNT_TYPE;
CREATE TYPE ACCOUNT_TYPE AS ENUM ('USER', 'SYNDICATE');

CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type ACCOUNT_TYPE NOT NULL DEFAULT 'USER',
    user_id uuid NOT NULL REFERENCES users (id),
    sups NUMERIC(28) NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

INSERT INTO accounts (user_id, sups)
SELECT id, sups from users WHERE id NOT IN (
    '2fa1a63e-a4fa-4618-921f-4b4d28132069',
    'ebf30ca0-875b-4e84-9a78-0b3fa36a1f87',
    '4fae8fdf-584f-46bb-9cb9-bb32ae20177e',
    '87c60803-b051-4abb-aa60-487104946bd7',
    'c579bb47-7efb-4286-a5cc-e5edbb54626d',
    '1a657a32-778e-4612-8cc1-14e360665f2b',
    '305da475-53dc-4973-8d78-a30d390d3de5',
    '15f29ee9-e834-4f76-aff8-31e39faabe2d',
    '2fa1a63e-a4fa-4618-921f-4b4d28132069',
    '1429a004-84a1-11ec-a8a3-0242ac120002');

-- XsynTreasuryUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT 'f16e175e-769d-4cfc-9c17-acdd433ecaf2', id, sups from users WHERE id = 'ebf30ca0-875b-4e84-9a78-0b3fa36a1f87';

-- SupremacyGameUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT 'e7b5b572-90ee-461b-bd12-b73a90f96cbf', id, sups from users WHERE id = '4fae8fdf-584f-46bb-9cb9-bb32ae20177e';

-- SupremacyBattleUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT 'ca9cac28-986b-409c-a9a1-6f0d4d86c125', id, sups from users WHERE id = '87c60803-b051-4abb-aa60-487104946bd7';

-- SupremacySupPoolUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT 'dd686216-d3b9-4f6d-b33d-a03756208bfa', id, sups from users WHERE id = 'c579bb47-7efb-4286-a5cc-e5edbb54626d';

-- SupremacyZaibatsuUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT '8d3c2947-0c03-4d42-bf36-023abff2ffe0', id, sups from users WHERE id = '1a657a32-778e-4612-8cc1-14e360665f2b';

-- SupremacyRedMountainUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT 'e21cb722-f01f-4690-ae85-7903e96bc52b', id, sups from users WHERE id = '305da475-53dc-4973-8d78-a30d390d3de5';

-- SupremacyBostonCyberneticsUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT '96d889e6-826b-4f21-941a-b3fdf063637a', id, sups from users WHERE id = '15f29ee9-e834-4f76-aff8-31e39faabe2d';

-- OnChainUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT '0a17f4af-afff-4ff4-8410-0e53f84e2ef9', id, sups from users WHERE id = '2fa1a63e-a4fa-4618-921f-4b4d28132069';

-- XsynSaleUserAccountID
INSERT INTO accounts (id, user_id, sups)
SELECT '572f781a-50e5-4ab4-98b5-33d6fd1b5dbe', id, sups from users WHERE id = '1429a004-84a1-11ec-a8a3-0242ac120002';


ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_id uuid REFERENCES accounts (id);

UPDATE
    users u
SET
    account_id = (SELECT a.id FROM accounts a WHERE a.user_id = u.id);

ALTER TABLE users
    ALTER COLUMN account_id SET NOT NULL;

ALTER TABLE accounts
    ALTER COLUMN type DROP DEFAULT,
    DROP COLUMN user_id;

ALTER TABLE users
    DROP COLUMN sups;

CREATE TABLE syndicates(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    founded_by_id uuid not null references users(id),
    name text NOT NULL,
    account_id uuid not null references accounts(id),
    created_at timestamptz not null default NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS TRANSACTION_DIRECTION;
CREATE TYPE TRANSACTION_DIRECTION AS ENUM ('P2P', 'P2S', 'S2P', 'S2S');

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS direction TRANSACTION_DIRECTION NOT NULL DEFAULT 'P2P',
    ADD COLUMN IF NOT EXISTS debit_account_id uuid REFERENCES accounts (id),
    ADD COLUMN IF NOT EXISTS credit_account_id uuid REFERENCES accounts (id);

UPDATE
    transactions t
SET
    debit_account_id = (SELECT u.account_id FROM users u where u.id = t.debit ),
    credit_account_id = (SELECT u.account_id FROM users u where u.id = t.credit );


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
    SELECT NEW.debit_account_id = '0a17f4af-afff-4ff4-8410-0e53f84e2ef9' OR (
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
    ALTER COLUMN direction DROP DEFAULT,
    ALTER COLUMN debit_account_id SET NOT NULL,
    ALTER COLUMN credit_account_id SET NOT NULL,
    DROP COLUMN IF EXISTS debit,
    DROP COLUMN IF EXISTS credit;

CREATE INDEX idx_transactions_created_at_descending ON transactions(created_at DESC);
CREATE INDEX idx_transactions_credit ON transactions(credit_account_id);
CREATE INDEX idx_transactions_debit ON transactions(debit_account_id);
CREATE INDEX idx_transactions_debit_credit ON transactions(debit_account_id,credit_account_id);
CREATE INDEX idx_transactions_credit_debit ON transactions(credit_account_id,debit_account_id);