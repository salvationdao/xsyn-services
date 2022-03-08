CREATE TABLE pending_refund (
    id                      UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    refunded_at             TIMESTAMPTZ NOT NULL,
    is_refunded             BOOLEAN NOT NULL DEFAULT false,
    refund_canceled_at      TIMESTAMPTZ,
    tx_hash                  TEXT NOT NULL DEFAULT '',
    transaction_reference   TEXT NOT NULL REFERENCES transactions(transaction_reference),
    deleted_at              TIMESTAMPTZ,
    updated_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);
