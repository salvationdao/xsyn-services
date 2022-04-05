ALTER TABLE state ADD COLUMN id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE state ALTER COLUMN sup_to_usd SET NOT NULL;

CREATE TABLE store_items (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id),
    faction_id UUID NOT NULL REFERENCES factions(id),
    usd_cent_cost INT NOT NULL CHECK (usd_cent_cost > 0),
    amount_sold INTEGER NOT NULL CHECK (amount_sold >= 0 AND amount_sold <= amount_available),
    amount_available INTEGER NOT NULL CHECK (amount_available >= 0),
    restriction_group TEXT NOT NULL CHECK (restriction_group IN ('NONE', 'WHITELIST', 'LOOTBOX', 'PRIZE')),
    is_default BOOLEAN NOT NULL,
    tier TEXT NOT NULL CHECK (tier IN ('MEGA', 'COLOSSAL', 'RARE', 'LEGENDARY', 'ELITE_LEGENDARY', 'ULTRA_RARE', 'EXOTIC', 'GUARDIAN', 'MYTHIC', 'DEUS_EX')),
    data JSONB NOT NULL,

    deleted_at TIMESTAMPTZ,
    refreshes_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchased_items (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id),
    store_item_id UUID NOT NULL REFERENCES store_items(id),
    external_token_id INTEGER NOT NULL,
    is_default BOOLEAN NOT NULL,
    tier TEXT NOT NULL CHECK (tier IN ('MEGA', 'COLOSSAL', 'RARE', 'LEGENDARY', 'ELITE_LEGENDARY', 'ULTRA_RARE', 'EXOTIC', 'GUARDIAN', 'MYTHIC', 'DEUS_EX')),
    hash TEXT UNIQUE NOT NULL CHECK (hash != ''),
    owner_id UUID NOT NULL REFERENCES users(id),
    data JSONB NOT NULL,
    UNIQUE (collection_id, external_token_id),

    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    minted_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    refreshes_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE item_onchain_transactions (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id),
    external_token_id INTEGER NOT NULL,
    tx_id TEXT NOT NULL,
    contract_addr TEXT NOT NULL,
    from_addr TEXT NOT NULL,
    to_addr TEXT NOT NULL,
    block_number INTEGER NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON item_onchain_transactions (to_addr);
CREATE INDEX ON item_onchain_transactions (external_token_id);
CREATE INDEX ON item_onchain_transactions (from_addr);
CREATE INDEX ON item_onchain_transactions (to_addr);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type TEXT NOT NULL CHECK (type IN ('ADMIN', 'MEMBER', 'GUEST')),
    
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
