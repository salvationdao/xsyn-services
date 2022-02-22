ALTER TABLE xsyn_metadata
    ALTER COLUMN image DROP NOT NULL,
    ALTER COLUMN image DROP DEFAULT,
    ALTER COLUMN animation_url DROP NOT NULL,
    ALTER COLUMN animation_url DROP DEFAULT;

