-- UPDATING Red Mountain store items - meta - Max Structure Hit Points to 1500 from 1000
WITH item AS (
    SELECT ('{'||pos-1||', "value"}')::TEXT[] AS path, id
    FROM xsyn_store,
         JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem ->> 'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1500', FALSE)
FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mega"}]' -- rarity we want to update
AND faction_id = (SELECT id FROM factions WHERE factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update


-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1530
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1500', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update
