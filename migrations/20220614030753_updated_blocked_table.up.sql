ALTER TABLE block_withdraw DROP CONSTRAINT block_withdraw_pkey;
ALTER TABLE block_withdraw
    ADD CONSTRAINT unique_address UNIQUE (public_address),
    ALTER COLUMN public_address SET NOT NULL;


ALTER TABLE block_withdraw
    ADD COLUMN id UUID DEFAULT gen_random_uuid(),
    ADD COLUMN block_sups_withdraws TIMESTAMPTZ NOT NULL DEFAULT '22 February 2024 00:00:00 GMT+08:00',
    ADD COLUMN block_nft_withdraws TIMESTAMPTZ NOT NULL DEFAULT '22 February 2024 00:00:00 GMT+08:00'
;

UPDATE block_withdraw SET id = gen_random_uuid() WHERE id IS NULL;

ALTER TABLE block_withdraw
    ADD CONSTRAINT pk_id PRIMARY KEY (id);