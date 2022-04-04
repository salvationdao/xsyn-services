BEGIN;

CREATE TABLE deposit_transactions
(
    id         UUID        PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id    UUID        REFERENCES users (id) NOT NULL,
    tx_hash    TEXT        UNIQUE      NOT NULL,
    amount     NUMERIC(28)             NOT NULL,
    status     TEXT                    NOT NULL CHECK (status IN ('pending', 'confirmed')) DEFAULT 'pending',
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ             NOT NULL DEFAULT NOW()
);

COMMIT;