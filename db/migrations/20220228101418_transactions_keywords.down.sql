BEGIN;

ALTER TABLE transactions
    ALTER COLUMN group_id DROP NOT NULL,
    ALTER COLUMN group_id DROP DEFAULT;

UPDATE transactions
    SET group_id = null
    WHERE group_id = '';

COMMIT;
