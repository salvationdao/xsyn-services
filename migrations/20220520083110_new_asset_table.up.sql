ALTER TABLE purchased_items
    RENAME TO purchased_items_old;

DROP TYPE IF EXISTS SERVICE;
CREATE TYPE SERVICE AS ENUM ('Supremacy', 'XSYN');


CREATE TABLE user_assets
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_id      UUID             NOT NULL REFERENCES collections (id),
    token_id           BIGINT           NOT NULL,
    tier               TEXT             NOT NULL,
    hash               TEXT             NOT NULL UNIQUE,
    owner_id           UUID             NOT NULL REFERENCES users,
    data               JSONB            NOT NULL DEFAULT '{}',
    attributes         JSONB            NOT NULL DEFAULT '{}',
    name               TEXT             NOT NULL,
    image_url          TEXT,
    external_url       TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    description        TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT,
    unlocked_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    minted_at          TIMESTAMPTZ,
    on_chain_status    TEXT             NOT NULL DEFAULT 'MINTABLE' CHECK (on_chain_status IN ('MINTABLE', 'STAKABLE', 'UNSTAKABLE')),
    deleted_at         TIMESTAMPTZ,
    data_refreshed_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    locked_to_service  UUID REFERENCES users (id),
    UNIQUE (collection_id, token_id)
);
