ALTER TABLE xsyn_metadata
    ALTER COLUMN image SET DEFAULT '',
    ALTER COLUMN image SET NOT NULL,
    ALTER COLUMN animation_url SET DEFAULT '',
    ALTER COLUMN animation_url SET NOT NULL;
