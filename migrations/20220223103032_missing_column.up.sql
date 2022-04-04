ALTER TABLE xsyn_store ADD COLUMN restriction TEXT NOT NULL NOT NULL CHECK (restriction IN ('', 'WHTIELIST', 'LOOTBOX' )) DEFAULT '';

