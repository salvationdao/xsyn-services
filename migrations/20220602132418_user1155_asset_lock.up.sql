CREATE TABLE asset1155_service_transfer_events
(
    id                BIGSERIAL PRIMARY KEY,
    user1155_asset_id UUID        NOT NULL REFERENCES user_assets_1155 (id),
    user_id           UUID        NOT NULL REFERENCES users (id),
    initiated_from    TEXT        NOT NULL DEFAULT 'XSYN',
    amount            INT         NOT NULL,
    CHECK ( amount > 0 ),
    from_service      UUID        NULL REFERENCES users (id),
    to_service        UUID REFERENCES users (id),
    transfer_tx_id    TEXT        NOT NULL REFERENCES transactions (id),
    transferred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

alter table user_assets_1155
    drop constraint if exists user_assets_1155_owner_id_collection_id_external_token_id_key;

CREATE TABLE deposit_asset1155_transactions
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id),
    collection_slug TEXT        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'pending',
    tx_hash         TEXT        NOT NULL,
    token_id        INT         NOT NULL,
    amount          INT         NOT NULL,
    CHECK ( amount > 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NULL
);
