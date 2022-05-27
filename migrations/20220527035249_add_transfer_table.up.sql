CREATE TABLE asset_service_transfer_events
(
    id                    UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_asset_id         UUID             NOT NULL REFERENCES user_assets,
    user_id               UUID             NOT NULL REFERENCES users (id),
    initiated_from        TEXT             NOT NULL DEFAULT 'XSYN',
    from_service          UUID REFERENCES users (id),
    to_service            UUID REFERENCES users (id),
    transfer_tx_id        TEXT             NOT NULL REFERENCES transactions (id),
    transfer_refund_tx_id TEXT REFERENCES transactions (id),
    transferred_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    refunded_at           TIMESTAMPTZ
);
