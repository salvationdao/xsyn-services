ALTER TABLE transactions
    RENAME TO transactions_old;

CREATE TABLE transactions
(
    id                     TEXT                      NOT NULL PRIMARY KEY,
    description            TEXT        DEFAULT ''    NOT NULL,
    transaction_reference  TEXT        DEFAULT ''    NOT NULL UNIQUE,
    amount                 NUMERIC(28)               NOT NULL CHECK (amount > 0.0),
    credit                 UUID                      NOT NULL REFERENCES users,
    debit                  UUID                      NOT NULL REFERENCES users,
    reason                 TEXT        DEFAULT '',
    created_at             TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    "group"                TEXT        DEFAULT ''    NOT NULL,
    sub_group              TEXT        DEFAULT '',
    related_transaction_id TEXT REFERENCES transactions,
    service_id             UUID REFERENCES users
);

CREATE INDEX transactions_credit_idx ON transactions (credit);
CREATE INDEX transactions_debit_idx ON transactions (debit);
CREATE INDEX group_idx ON transactions ("group");
CREATE INDEX sub_group_idx ON transactions (sub_group);
CREATE INDEX idx_transactions_created_at ON transactions (created_at);
CREATE INDEX idx_transactions_created_desc_at ON transactions (created_at DESC);
CREATE INDEX idx_user_search ON transactions (credit, debit, created_at DESC);

CREATE TRIGGER t_transactions_insert
    BEFORE INSERT OR UPDATE
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE uppercase_group_and_sub_group();

CREATE TRIGGER trigger_check_balance
    BEFORE INSERT
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE check_balances();

-- Finally copy the data to our new partitioned table
INSERT INTO transactions (id, description, transaction_reference, amount, credit, debit, created_at, "group", sub_group, service_id,
                          related_transaction_id)
SELECT id,
       description,
       transaction_reference,
       amount,
       credit,
       debit,
       created_at,
       "group",
       sub_group,
       service_id,
       related_transaction_id
FROM transactions_old
WHERE created_at > NOW() - INTERVAL '1 month';

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
