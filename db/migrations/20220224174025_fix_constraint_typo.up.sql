ALTER TABLE xsyn_store DROP CONSTRAINT IF EXISTS xsyn_store_restriction_check;
UPDATE xsyn_store SET restriction = 'WHITELIST' WHERE restriction = 'WHTIELIST';
ALTER TABLE xsyn_store ADD CONSTRAINT xsyn_store_restriction_check CHECK (restriction IN ('', 'WHITELIST', 'LOOTBOX' ));
