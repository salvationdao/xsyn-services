/**
  Add shield recharge stat to all mechs on the store
 **/

UPDATE xsyn_store
SET attributes = attributes ||'[{"trait_type": "Shield Recharge Rate", "value": 80, "display_type": "number"}]'::jsonb
WHERE xsyn_store.attributes @> '[{"trait_type": "Asset Type", "value": "War Machine"}]'; -- rarity we want to update
-- "display_type": "number"

----------------------  Update red mountain mechs  --------------------------------

-- UPDATING Red Mountain store items - Mega - Max Structure Hit Points to 1500 from 1000
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
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1530', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1560
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1560', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1590
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1590', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1620
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1620', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1650
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1650', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1680
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1650', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1710
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1710', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1740
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1740', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1800
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '1800', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Red Mountain Offworld Mining Corporation'); -- faction we want to update

---------------------- Zaibatsu mech updates ----------------------------------

-- UPDATING Zaibatsu Heavy Industries - Mega - Shield Recharge Rate to 100
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '100', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
  AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mega"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Colossal - Shield Recharge Rate to 102
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '102', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Rare - Shield Recharge Rate to 104
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '104', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Legendary - Shield Recharge Rate to 106
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '106', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Elite Legendary - Shield Recharge Rate to 108
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '108', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Ultra Rare - Shield Recharge Rate to 110
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '110', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Exotic - Shield Recharge Rate to 112
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '112', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Guardian - Shield Recharge Rate to 114
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '114', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

-- UPDATING Zaibatsu Heavy Industries - Mythic - Shield Recharge Rate to 116
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '116', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update


-- UPDATING Zaibatsu Heavy Industries - Deus ex - Shield Recharge Rate to 120
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '120', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Zaibatsu Heavy Industries'); -- faction we want to update

---------------   Boston mech updates  --------------------------

-- UPDATING Boston Cybernetics - Mega - Speed to 2750
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '2750', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
  AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mega"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Colossal - Speed to 2,805
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '2805', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Rare - Speed to 2860
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '2860', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Legendary - Speed to 2915
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '2915', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Elite Legendary - Speed to 2970
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '2970', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Ultra Rare - Speed to 3025
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '3025', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Exotic - Speed to 3080
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '3080', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Guardian - Speed to 3135
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '3135', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Mythic - Speed to 3190
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '3190', FALSE)
    FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update

-- UPDATING Boston Cybernetics - Deus ex - Speed to 3300
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, id
    FROM xsyn_store, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
    )
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '3300', FALSE)
FROM item
WHERE item.id = xsyn_store.id
AND xsyn_store.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
AND faction_id = (SELECT id FROM factions where factions.label = 'Boston Cybernetics'); -- faction we want to update
