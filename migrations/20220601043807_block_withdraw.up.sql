CREATE TABLE block_withdraw
(
    public_address TEXT PRIMARY KEY NOT NULL
);

ALTER TABLE block_withdraw
    ADD COLUMN note       TEXT,
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
