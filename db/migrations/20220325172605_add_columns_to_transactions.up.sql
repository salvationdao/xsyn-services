ALTER TABLE transactions ADD COLUMN related_transaction_id TEXT REFERENCES transactions(id);
ALTER TABLE transactions ADD COLUMN service_id uuid REFERENCES users(id);

-- update all transaction that have battle in them to have the service id of the supremacy user
UPDATE transactions SET service_id = '4fae8fdf-584f-46bb-9cb9-bb32ae20177e' WHERE description ILIKE '%battle%' OR description ILIKE '%queue%';
