ALTER TABLE collections
    ADD COLUMN background_url TEXT,
    ADD COLUMN logo_url       TEXT,
    ADD COLUMN description    TEXT,
    ADD COLUMN external_token_ids BIGINT[],
    ADD COLUMN transfer_contract TEXT;

UPDATE collections
SET logo_url       = 'https://afiles.ninja-cdn.com/passport/collections/supremacy-achievements/logo.png',
    background_url = 'https://afiles.ninja-cdn.com/passport/collections/supremacy-achievements/background.png',
    is_visible     = true,
    description    = 'Supremacy is a collection of games that connect players into one immersive, interactive, and interconnected universe offering a unique experience to those looking for a Metaverse that blurs the line between the real world and the game world.',
    external_token_ids = ARRAY[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ,11, 12, 13, 14, 15, 16, 17, 18, 19],
    transfer_contract = '0x52b38626D3167e5357FE7348624352B7062fE271'
WHERE slug = 'supremacy-achievements';

ALTER TABLE user_assets_1155
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN service_id UUID NULL,
    ADD CONSTRAINT check_count CHECK(count >= 0);