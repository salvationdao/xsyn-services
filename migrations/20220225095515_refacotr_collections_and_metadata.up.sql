
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

ALTER TABLE users ADD COLUMN metadata JSONB DEFAULT '{}';
