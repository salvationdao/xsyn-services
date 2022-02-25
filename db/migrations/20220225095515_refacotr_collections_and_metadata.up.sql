
--  update collections
ALTER TABLE collections ADD COLUMN mint_contract TEXT DEFAULT ''; -- Collection table should have contract address to mint against
ALTER TABLE collections ADD COLUMN stake_contract TEXT DEFAULT ''; -- Collection table should have contract address to mint against
UPDATE collections set mint_contract = '0x52d7A31e2f5CfA6De6BABb2787c0dF842298f5e6', stake_contract = '0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95'  WHERE name = 'Supremacy';
UPDATE collections set mint_contract = '0x651D4424F34e6e918D8e4D2Da4dF3DEbDAe83D0C', stake_contract = '0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95' WHERE name = 'Supremacy Genesis';

-- drop unneeded table
DROP TABLE IF EXISTS war_machine_ability_sups_cost;

-- inser AI collection
INSERT INTO collections (id, name)
VALUES ('9cdf55aa-217b-4821-aa77-bc8555195f23', 'Supremacy AI');

-- clear func
CREATE OR REPLACE FUNCTION updateXsyn_metadataKeywords()
    RETURNS TRIGGER
AS
$updateXsyn_metadataKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    RETURN NULL;
END;
$updateXsyn_metadataKeywords$
LANGUAGE plpgsql;

-- delete current data
DELETE FROM xsyn_assets; -- CLEAR XSYN_ASSET DATA
DELETE FROM xsyn_metadata; -- CLEAR XSYN_METADATA

-- update xsyn metadata table
ALTER TABLE xsyn_metadata DROP CONSTRAINT xsyn_metadata_collection_id_fkey; -- drop metadata PK
ALTER TABLE xsyn_assets DROP COLUMN token_id; -- drop this before xsyn_metadata id column
ALTER TABLE xsyn_metadata DROP COLUMN token_id; -- id column
ALTER TABLE xsyn_metadata ADD COLUMN hash TEXT PRIMARY KEY UNIQUE; -- add new PK column (will be hash of collectio + token id)
ALTER TABLE xsyn_metadata ADD COLUMN external_token_id NUMERIC(78, 0); -- add new token id column
ALTER TABLE xsyn_metadata ADD CONSTRAINT xsyn_metadata_token_collection_unique UNIQUE(external_token_id, collection_id);

-- update xsyn assets table
-- ALTER TABLE xsyn_assets DROP CONSTRAINT xsyn_assets_token_id_fkey; -- drop token id PK

ALTER TABLE xsyn_assets ADD COLUMN external_token_id NUMERIC(78, 0); -- add new token id column
ALTER TABLE xsyn_assets ADD COLUMN collection_id UUID REFERENCES collections (id); -- add FK to collections
ALTER TABLE xsyn_assets ADD COLUMN metadata_hash TEXT PRIMARY KEY REFERENCES xsyn_metadata(hash);
ALTER TABLE xsyn_assets ADD COLUMN signature_expiry TEXT DEFAULT '';

-- remake func
CREATE OR REPLACE FUNCTION updateXsyn_metadataKeywords()
    RETURNS TRIGGER
AS
$updateXsyn_metadataKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.external_url), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.image), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            xsyn_metadata
        SET keywords = temp
        WHERE hash = NEW.hash;
    END IF;
    RETURN NULL;
END;
$updateXsyn_metadataKeywords$
LANGUAGE plpgsql;

-- add fk contraincts to asset table


-- add default mechs metadata to the Supremacy AI collection
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('ZXga92AmGD', 1,  '9cdf55aa-217b-4821-aa77-bc8555195f23',  'https://passport.xsyn.io/asset/ZXga92AmGD', '', '', '[{"value": "Zaibatsu", "trait_type": "Brand"}, {"value": "WREX", "trait_type": "Model"}, {"value": "Black", "trait_type": "Skin"}, {"value": "Alex", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2500, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Sniper Rifle", "trait_type": "Weapon One"}, {"value": "Laser Sword", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('dbYaD4a0Zj', 2,  '9cdf55aa-217b-4821-aa77-bc8555195f23',  'https://passport.xsyn.io/asset/dbYaD4a0Zj', '', '', '[{"value": "Zaibatsu", "trait_type": "Brand"}, {"value": "WREX", "trait_type": "Model"}, {"value": "Black", "trait_type": "Skin"}, {"value": "John", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2500, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Sniper Rifle", "trait_type": "Weapon One"}, {"value": "Laser Sword", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('l7epj2pPL4', 3, '9cdf55aa-217b-4821-aa77-bc8555195f23',   'https://passport.xsyn.io/asset/l7epj2pPL4', '', '', '[{"value": "Zaibatsu", "trait_type": "Brand"}, {"value": "WREX", "trait_type": "Model"}, {"value": "Black", "trait_type": "Skin"}, {"value": "Mac", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2500, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Sniper Rifle", "trait_type": "Weapon One"}, {"value": "Laser Sword", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('kN7aVgAenK', 4, '9cdf55aa-217b-4821-aa77-bc8555195f23',   'https://passport.xsyn.io/asset/kN7aVgAenK', '', '', '[{"value": "Red Mountain", "trait_type": "Brand"}, {"value": "BXSD", "trait_type": "Model"}, {"value": "Red_Steel", "trait_type": "Skin"}, {"value": "Vinnie", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1500, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 1750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Auto Cannon", "trait_type": "Weapon One"}, {"value": "Auto Cannon", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('wdBAN1aeo5', 5,  '9cdf55aa-217b-4821-aa77-bc8555195f23',  'https://passport.xsyn.io/asset/wdBAN1aeo5', '', '', '[{"value": "Red Mountain", "trait_type": "Brand"}, {"value": "BXSD", "trait_type": "Model"}, {"value": "Pink", "trait_type": "Skin"}, {"value": "Owen", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1500, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 1750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Auto Cannon", "trait_type": "Weapon One"}, {"value": "Auto Cannon", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('018pkXaRWM', 6, '9cdf55aa-217b-4821-aa77-bc8555195f23',   'https://passport.xsyn.io/asset/018pkXaRWM', '', '', '[{"value": "Red Mountain", "trait_type": "Brand"}, {"value": "BXSD", "trait_type": "Model"}, {"value": "Pink", "trait_type": "Skin"}, {"value": "James", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1500, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 1750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 2, "trait_type": "Turret Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Auto Cannon", "trait_type": "Weapon One"}, {"value": "Auto Cannon", "trait_type": "Weapon Two"}, {"value": "Rocket Pod", "trait_type": "Turret One"}, {"value": "Rocket Pod", "trait_type": "Turret Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('B8x3qdAy6K', 7,  '9cdf55aa-217b-4821-aa77-bc8555195f23',  'https://passport.xsyn.io/asset/B8x3qdAy6K', '', '', '[{"value": "Boston Cybernetics", "trait_type": "Brand"}, {"value": "XFVS", "trait_type": "Model"}, {"value": "BlueWhite", "trait_type": "Skin"}, {"value": "Darren", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Plasma Rifle", "trait_type": "Weapon One"}, {"value": "Sword", "trait_type": "Weapon Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('D16aRep0Zo', 8,  '9cdf55aa-217b-4821-aa77-bc8555195f23',  'https://passport.xsyn.io/asset/D16aRep0Zo', '', '', '[{"value": "Boston Cybernetics", "trait_type": "Brand"}, {"value": "XFVS", "trait_type": "Model"}, {"value": "Police_DarkBlue", "trait_type": "Skin"}, {"value": "Yong", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Plasma Rifle", "trait_type": "Weapon One"}, {"value": "Sword", "trait_type": "Weapon Two"}, {"value": "Shield", "trait_type": "Utility One"}]');
INSERT INTO public.xsyn_metadata (hash, external_token_id, collection_id, external_url, image, animation_url, attributes) VALUES ('4Q1p8dpqwX', 9, '9cdf55aa-217b-4821-aa77-bc8555195f23',   'https://passport.xsyn.io/asset/4Q1p8dpqwX', '', '', '[{"value": "Boston Cybernetics", "trait_type": "Brand"}, {"value": "XFVS", "trait_type": "Model"}, {"value": "Police_DarkBlue", "trait_type": "Skin"}, {"value": "Corey", "trait_type": "Name"}, {"value": "War Machine", "trait_type": "Asset Type"}, {"value": 1000, "trait_type": "Max Structure Hit Points", "display_type": "number"}, {"value": 1000, "trait_type": "Max Shield Hit Points", "display_type": "number"}, {"value": 2750, "trait_type": "Speed", "display_type": "number"}, {"value": 2, "trait_type": "Weapon Hardpoints", "display_type": "number"}, {"value": 1, "trait_type": "Utility Slots", "display_type": "number"}, {"value": "Plasma Rifle", "trait_type": "Weapon One"}, {"value": "Sword", "trait_type": "Weapon Two"}, {"value": "Shield", "trait_type": "Utility One"}]');

update xsyn_store set usd_cent_cost = 100 where attributes @> '[{"trait_type": "Rarity"}]' and attributes @> '[{"value": "Mega"}]';

UPDATE xsyn_store SET restriction = 'WHITELIST' WHERE description = 'Gold';


update xsyn_store set collection_id = (SELECT id from collections c where c.name = 'Supremacy Genesis')
where attributes @> '[{"trait_type": "Rarity"}]' and attributes @> '[{"value": "Mega"}]';

ALTER TABLE users ADD COLUMN metadata JSONB DEFAULT '';
