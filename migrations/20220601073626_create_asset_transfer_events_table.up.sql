CREATE TABLE asset_transfer_events
(
    id              BIGSERIAL PRIMARY KEY,
    user_asset_id   UUID        NOT NULL REFERENCES user_assets,
    user_asset_hash TEXT        NOT NULL REFERENCES user_assets (hash),
    from_user_id    UUID        NOT NULL REFERENCES users (id),
    to_user_id      UUID        NOT NULL REFERENCES users (id),
    initiated_from  TEXT        NOT NULL DEFAULT 'XSYN',
    transfer_tx_id  TEXT REFERENCES transactions (id),
    transferred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
