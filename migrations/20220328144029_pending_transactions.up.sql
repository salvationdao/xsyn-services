BEGIN;

ALTER TABLE pending_refund
    ADD COLUMN withdraw_transaction_id TEXT REFERENCES transactions (id),
    ADD COLUMN reversal_transaction_id TEXT REFERENCES transactions (id);

COMMIT;
