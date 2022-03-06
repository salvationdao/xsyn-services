CREATE TABLE store_items (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id),
    faction_id UUID NOT NULL REFERENCES factions(id),
    usd_cent_cost INT NOT NULL CHECK (usd_cent_cost > 0),
    amount_sold INTEGER NOT NULL CHECK (amount_sold >= 0),
    amount_available INTEGER NOT NULL CHECK (amount_available >= 0),
    restriction_group TEXT NOT NULL CHECK (restriction_group IN ('NONE', 'WHITELIST', 'LOOTBOX')),
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
    hash TEXT UNIQUE NOT NULL CHECK (hash != ''),
    owner_id UUID NOT NULL REFERENCES users(id),
    data JSONB NOT NULL,

    minted_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    refreshes_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE item_onchain_transactions (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    purchased_item_id UUID NOT NULL REFERENCES purchased_items(id),
    tx_id TEXT NOT NULL,
    contract_addr TEXT NOT NULL,
    from_addr TEXT NOT NULL,
    to_addr TEXT NOT NULL,
    
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
