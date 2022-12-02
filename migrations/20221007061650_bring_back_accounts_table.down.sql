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

CREATE INDEX transactions_credit_idx ON transactions (credit);
CREATE INDEX transactions_debit_idx ON transactions (debit);

UPDATE
    transactions t
SET debit = t.debit_account_id,
    credit = t.credit_account_id;

ALTER TABLE transactions
    ALTER COLUMN debit SET NOT NULL,
    ALTER COLUMN credit SET NOT NULL;

ALTER TABLE users
    ADD COLUMN sups  NUMERIC(28)  NOT NULL DEFAULT 0 CHECK (sups >= 0 OR id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' );

UPDATE users SET sups = (SELECT sups FROM accounts WHERE accounts.id = users.id);

-- rewrite trigger
CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughfunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF new.debit = new.credit THEN
        RAISE EXCEPTION 'unable to transfer to self';
    END IF;

    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT new.debit = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (SELECT accounts.sups >= new.amount
                                                                  FROM accounts
                                                                  WHERE accounts.id = new.debit)
    INTO enoughfunds;
    -- if enough funds then make the updates to the user table
    IF enoughfunds THEN
        UPDATE users SET sups = sups - new.amount WHERE users.id = new.debit;
        UPDATE users SET sups = sups + new.amount WHERE users.id = new.credit;
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
SELECT transactions.credit,
       transactions.id,
       transactions.amount
FROM transactions
UNION ALL
SELECT transactions.debit,
       transactions.id,
       (0.0 - transactions.amount)
FROM transactions;

ALTER TABLE transactions
    DROP COLUMN debit_account_id,
    DROP COLUMN credit_account_id;



ALTER TABLE users
    DROP COLUMN account_id;

DROP TABLE syndicates;
DROP TABLE accounts;

DROP TYPE IF EXISTS ACCOUNT_TYPE;
