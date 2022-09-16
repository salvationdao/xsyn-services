ALTER TABLE transactions
    RENAME TO transactions_old;

CREATE TABLE transactions_master
(
    id                     TEXT                      NOT NULL,
    description            TEXT        DEFAULT ''    NOT NULL,
    transaction_reference  TEXT        DEFAULT ''    NOT NULL,
    amount                 NUMERIC(28)               NOT NULL CHECK (amount > 0.0),
    credit                 UUID                      NOT NULL REFERENCES users (id),
    debit                  UUID                      NOT NULL REFERENCES users (id),
    reason                 TEXT        DEFAULT '',
    created_at             TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    "group"                TEXT        DEFAULT ''    NOT NULL,
    sub_group              TEXT        DEFAULT '',
    related_transaction_id TEXT,
    service_id             UUID REFERENCES users (id)
) PARTITION BY RANGE (created_at);

CREATE INDEX transactions_master_idx ON ONLY transactions_master (id);

-- reference https://stackoverflow.com/questions/53600144/how-to-migrate-an-existing-postgres-table-to-partitioned-table-as-transparently
CREATE OR REPLACE FUNCTION createpartitionifnotexists(created_at TIMESTAMPTZ) RETURNS VOID
AS
$body$
DECLARE
    monthstart                DATE := DATE_TRUNC('month', created_at);
    DECLARE monthendexclusive DATE := monthstart + INTERVAL '1 month';
    -- We infer the name of the table from the date that it should contain
    -- E.g. a date in June 2005 should be int the table transactions_200506:
    DECLARE tablename         TEXT := 'transactions_' || TO_CHAR(created_at, 'YYYYmm');
BEGIN
    -- Check if the table we need for the supplied date exists.
    -- If it does not exist...:
    IF TO_REGCLASS(tablename) IS NULL THEN
        -- Generate a new table that acts as a partition for transactions:
        EXECUTE FORMAT('create table %I partition of transactions_master for values from (%L) to (%L)', tablename, monthstart,
                       monthendexclusive);
        -- Unfortunatelly Postgres forces us to define index for each table individually:
        EXECUTE FORMAT('create unique index on %I (id, created_at)', tablename);
    END IF;
END;
$body$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_balance
    ON transactions_old;

-- DROP FUNCTION check_balances("from" UUID, "to" UUID, amount NUMERIC(28)) CASCADE;

CREATE OR REPLACE FUNCTION check_balances("from" UUID, "to" UUID, amount NUMERIC(28)) RETURNS VOID AS
$check_balances$
DECLARE
    enoughFunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF "from" = "to" THEN
        RAISE EXCEPTION 'unable to transfer to self';
    END IF;
    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT "from" = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (SELECT sups >= amount
                                                                  FROM users
                                                                  WHERE id = "from")
    INTO enoughFunds;
    -- if enough funds then make the updates to the user table
    IF enoughFunds THEN
        UPDATE users SET sups = sups + amount WHERE id = "to";
        UPDATE users SET sups = sups - amount WHERE id = "from";
        RETURN;
    ELSE
        RAISE EXCEPTION 'not enough funds';
    END IF;
END
$check_balances$
    LANGUAGE plpgsql;

CREATE OR REPLACE VIEW transactions AS
SELECT *
FROM transactions_master;

CREATE OR REPLACE RULE autocall_createpartitionifnotexists AS ON INSERT
    TO transactions
    DO INSTEAD (

    SELECT createpartitionifnotexists(NEW.created_at);
    SELECT check_balances(NEW.debit, NEW.credit, NEW.amount);
    INSERT INTO transactions_master (id, description, transaction_reference, amount, credit, debit, reason, created_at, "group", sub_group,
                                     related_transaction_id, service_id)
    VALUES (new.id, new.description, new.transaction_reference, new.amount, new.credit, new.debit, new.reason, new.created_at, new."group",
            new.sub_group, new.related_transaction_id, new.service_id)
    );

-- Finally copy the data to our new partitioned table
INSERT INTO transactions (id, description, transaction_reference, amount, credit, debit, reason, created_at, "group", sub_group,
                          related_transaction_id, service_id)
SELECT *
FROM transactions_old WHERE created_at > NOW() - INTERVAL '1 month';

ALTER TABLE asset_transfer_events
    DROP CONSTRAINT asset_transfer_events_transfer_tx_id_fkey;

ALTER TABLE asset_service_transfer_events
    DROP CONSTRAINT asset_service_transfer_events_transfer_tx_id_fkey;

ALTER TABLE asset1155_service_transfer_events
    DROP CONSTRAINT asset1155_service_transfer_events_transfer_tx_id_fkey;

ALTER TABLE pending_refund
    DROP CONSTRAINT pending_refund_reversal_transaction_id_fkey,
    DROP CONSTRAINT pending_refund_transaction_reference_fkey,
    DROP CONSTRAINT pending_refund_withdraw_transaction_id_fkey;
