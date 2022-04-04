BEGIN;

UPDATE transactions
    SET group_id = ''
    WHERE group_id IS NULL;

ALTER TABLE transactions
    ALTER COLUMN group_id SET NOT NULL,
    ALTER COLUMN group_id SET DEFAULT '';

COMMIT;
