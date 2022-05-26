DROP TYPE IF EXISTS CONTRACT_TYPE;
CREATE TYPE CONTRACT_TYPE AS ENUM ('ERC-721', 'EIP-1155');

ALTER TABLE collections
    ADD contract_type CONTRACT_TYPE;

INSERT INTO collections (id, name, logo_blob_id, slug, mint_contract, stake_contract, is_visible, contract_type)
VALUES ('8ccb689f-f6fe-43dd-92f5-dcd6e61b5614', 'Supremacy Achievements', null, 'supremacy-achievements',
        '0x17F5655c7D834e4772171F30E7315bbc3221F1eE', null,
        false, 'EIP-1155');

UPDATE collections
SET contract_type = 'ERC-721'
WHERE mint_contract = '0xCA949036Ad7cb19C53b54BdD7b358cD12Cf0b810'
   or mint_contract = '0x651D4424F34e6e918D8e4D2Da4dF3DEbDAe83D0C';

CREATE TABLE user_assets_1155
(
    id                UUID PRIMARY KEY NOT NULL default gen_random_uuid(),
    owner_id          UUID             NOT NULL REFERENCES users (id),
    collection_id     UUID             NOT NULL REFERENCES collections (id),
    external_token_id INT              NOT NULL,
    count             INT              NOT NULL default 0
        CHECK ( count > 0 ),
    label             TEXT             NOT NULL,
    description       TEXT             NOT NULL,
    image_url         TEXT             NOT NULL,
    animation_url     TEXT             NULL,
    keycard_group     TEXT             NOT NULL,
    attributes        jsonb            NOT NULL,
    service_id        UUID             NULL,
    created_at        TIMESTAMPTZ      NOT NULL DEFAULT now(),
    UNIQUE (owner_id, collection_id, external_token_id)
);
