CREATE TABLE pending_refund (
    id                      UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id),
    amount_sups             NUMERIC(28) NOT NULL CHECK (amount_sups > 0),
    refunded_at             TIMESTAMPTZ NOT NULL,
    is_refunded             BOOLEAN NOT NULL DEFAULT false,
    refund_canceled_at      TIMESTAMPTZ,
    tx_hash                 TEXT NOT NULL DEFAULT '',
    transaction_reference   TEXT NOT NULL REFERENCES transactions(transaction_reference),
    deleted_at              TIMESTAMPTZ,
    updated_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);
