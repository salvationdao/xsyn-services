CREATE TABLE pending_1155_rollback
(
    id                 UUID PRIMARY KEY NOT NULL default gen_random_uuid(),
    user_id            UUID             NOT NULL REFERENCES users (id),
    asset_id           UUID             NOT NULL REFERENCES user_assets_1155 (id),
    count              INT              NOT NULL DEFAULT 1
        CHECK ( count > 0),
    is_refunded        BOOLEAN          NOT NULL DEFAULT false,
    refunded_at        TIMESTAMPTZ      NOT NULL,
    refund_canceled_at TIMESTAMPTZ,
    tx_hash            TEXT             NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ      NULL
)