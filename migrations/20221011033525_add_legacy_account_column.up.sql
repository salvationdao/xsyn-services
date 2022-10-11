-- here we;
-- add new column `legacy_account_id` to users
-- add new temp column to accounts `user_id_temp`
-- set user_id_temp
-- move current account id to `legacy_account_id`
-- create a new row in accounts for each user
-- drop `user_id_temp` column
-- move transactions to transactions old
-- delete all the transactions from transactions since we just moved them to transactions_old
-- add new account type to allow a check for sups >= 0 or account type == on chain
-- drop constrain to allow the negative account transfer their sups
-- rewrite trigger to allow for new negative account (we just check the account id now)
-- move sups from legacy account to new account

-- add new column `legacy_account_id` to users
ALTER TABLE users
    ADD COLUMN legacy_account_id UUID REFERENCES accounts (id);

-- add new temp column to accounts `user_id_temp`
ALTER TABLE accounts
    ADD COLUMN user_id_temp UUID REFERENCES users (id);

-- set user_id_temp
UPDATE accounts a
SET user_id_temp = (SELECT id FROM users u WHERE account_id = a.id);

-- move current account id to `legacy_account_id`
UPDATE users
SET legacy_account_id = account_id;

-- create a new row in accounts for each user
WITH inserted_accounts AS (
    INSERT INTO accounts (type, sups, user_id_temp)
        SELECT la.type, 0, la.user_id_temp
        FROM users _u
                 INNER JOIN accounts la ON _u.legacy_account_id = la.id
        RETURNING user_id_temp, id)
UPDATE users u
SET account_id = inserted_accounts.id
FROM inserted_accounts
WHERE inserted_accounts.user_id_temp = u.id;

-- drop `user_id_temp` column
ALTER TABLE accounts
    DROP COLUMN user_id_temp;

-- move transactions to transactions old
-- we need to remove the trigger on insert to old transactions table so it doesn't update balance again, we're not making a transactions, we're just moving data.
DROP TRIGGER IF EXISTS  trigger_check_balance
    ON transactions_old;
INSERT INTO transactions_old(id, amount, credit, debit, related_transaction_id, service_id, created_at)
SELECT t.id, t.amount, t.credit_account_id, t.debit_account_id, t.related_transaction_id, t.service_id, t.created_at
FROM transactions t;

-- delete all the transactions from transactions since we just moved them to transactions_old
DELETE FROM transactions;

-- add new account type to allow a check for sups >= 0 or account type == on chain
BEGIN;
ALTER TYPE account_type ADD VALUE IF NOT EXISTS 'ONCHAIN';
COMMIT; -- cannot alter a type and use it in the same tx
UPDATE accounts SET "type" = 'ONCHAIN' WHERE id = (SELECT u.account_id FROM users u WHERE u.id = '2fa1a63e-a4fa-4618-921f-4b4d28132069');
UPDATE accounts SET "type" = 'ONCHAIN' WHERE id = (SELECT u.legacy_account_id FROM users u WHERE u.id = '2fa1a63e-a4fa-4618-921f-4b4d28132069');
ALTER TABLE accounts
    DROP CONSTRAINT IF EXISTS accounts_check;
ALTER TABLE accounts
    ADD CHECK (sups >= 0 OR "type" = 'ONCHAIN');

-- drop constrain to allow the negative account transfer their sups
ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_amount_check1;

-- rewrite trigger to allow for new negative account (we just check the account id now)
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
    SELECT (SELECT id = '2fa1a63e-a4fa-4618-921f-4b4d28132069'
            FROM users
            WHERE account_id = new.debit_account_id
               OR legacy_account_id = new.debit_account_id)
               OR (SELECT accounts.sups >= new.amount
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


-- move sups from legacy account to new account
WITH sups_to_move AS (SELECT _u.id AS user_id, _a.id AS new_id, _la.id AS old_id, _la.sups
                      FROM users _u
                               INNER JOIN accounts _a ON _u.account_id = _a.id
                               INNER JOIN accounts _la ON _u.legacy_account_id = _la.id
                      WHERE _la.sups > 0 OR _u.id = '2fa1a63e-a4fa-4618-921f-4b4d28132069'
)
INSERT
INTO transactions(id, transaction_reference, amount, debit_account_id, credit_account_id)
SELECT 'SUP MIGRATION - ' || stm.user_id, 'SUP MIGRATION - ' || stm.user_id, stm.sups, stm.old_id, stm.new_id
FROM sups_to_move stm;
