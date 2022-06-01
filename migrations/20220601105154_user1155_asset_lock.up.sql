CREATE TABLE asset1155_service_transfer_events
(
    id                BIGSERIAL PRIMARY KEY,
    user1155_asset_id UUID        NOT NULL REFERENCES user_assets_1155 (id),
    user_id           UUID        NOT NULL REFERENCES users (id),
    initiated_from    TEXT        NOT NULL DEFAULT 'XSYN',
    amount            INT         NOT NULL DEFAULT 1
        CHECK ( amount > 0 ),
    from_service      UUID        NULL REFERENCES users (id),
    to_service        UUID REFERENCES users (id),
    transfer_tx_id    TEXT        NOT NULL REFERENCES transactions (id),
    transferred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

alter table user_assets_1155
    drop constraint if exists user_assets_1155_owner_id_collection_id_external_token_id_key;