ALTER TABLE collections
    ALTER COLUMN mint_contract DROP DEFAULT;
ALTER TABLE purchased_items
    ALTER COLUMN store_item_id DROP NOT NULL;

INSERT INTO collections (name,
                         slug,
                         is_visible)
VALUES ('Supremacy General', 'supremacy-general', TRUE);

INSERT INTO collections (name,
                         slug,
                         is_visible)
VALUES ('Supremacy Consumables', 'supremacy-consumables', TRUE);
