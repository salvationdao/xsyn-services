CREATE TABLE asset_service_transfer_events
(
    id             BIGSERIAL PRIMARY KEY,
    user_asset_id  UUID        NOT NULL REFERENCES user_assets,
    user_id        UUID        NOT NULL REFERENCES users (id),
    initiated_from TEXT        NOT NULL DEFAULT 'XSYN',
    from_service   UUID REFERENCES users (id),
    to_service     UUID REFERENCES users (id),
    transfer_tx_id TEXT        NOT NULL REFERENCES transactions (id),
    transferred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE purchased_items_old
    DROP CONSTRAINT purchased_items_on_chain_status_check,
    ADD CONSTRAINT purchased_items_on_chain_status_check CHECK (on_chain_status IN
                                                                ('MINTABLE', 'STAKABLE', 'UNSTAKABLE',
                                                                 'UNSTAKABLE_OLD'));

UPDATE purchased_items_old
SET on_chain_status = 'UNSTAKABLE_OLD'
WHERE on_chain_status = 'UNSTAKABLE';
