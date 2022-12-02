CREATE TABLE user_asset_on_chain_status
(
    id              UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_hash      TEXT        NOT NULL REFERENCES user_assets (hash),
    collection_id   UUID        NOT NULL REFERENCES collections (id),
    on_chain_status TEXT        NOT NULL             DEFAULT 'MINTABLE' CHECK (on_chain_status IN
                                                                               ('MINTABLE', 'STAKABLE', 'UNSTAKABLE',
                                                                                'UNSTAKABLE_OLD')),
    updated_at      TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

INSERT INTO user_asset_on_chain_status(asset_hash, collection_id, on_chain_status)
SELECT hash, collection_id, on_chain_status
FROM user_assets;

-- USING THIS ONE IN DEV SO IT ERRORS WHERE IT WAS USED
-- ALTER TABLE user_assets
--     DROP COLUMN on_chain_status;

-- THIS IS THE ONE WILL USE
ALTER TABLE user_assets
    RENAME COLUMN on_chain_status TO  on_chain_status_old;
