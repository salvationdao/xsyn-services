/**
  Add shield recharge stat to all mechs on the metadata
 **/

UPDATE xsyn_metadata
SET attributes = attributes ||'[{"trait_type": "Shield Recharge Rate", "value": 80}]'::jsonb
WHERE xsyn_metadata.attributes @> '[{"trait_type": "Asset Type", "value": "War Machine", "display_type": "number"}]';

/**
  Update red mountain mechs
 **/

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1530
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1530', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1560
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1560', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1590
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1590', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1620
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1620', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1650
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1650', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1680
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1650', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update


-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1710
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1710', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1740
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1740', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Red Mountain - Colossal - Max Structure Hit Points to 1800
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Max Structure Hit Points' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '1800', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Red Mountain"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Olympus Mons LY07", "trait_type": "Model"}]'; -- mech model to update

/************************************
    Zaibatsu mech updates
**************************************/

-- UPDATING Zaibatsu Heavy Industries - Colossal - Shield Recharge Rate to 81.6
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '81.6', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Rare - Shield Recharge Rate to 83.2
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '83.2', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Legendary - Shield Recharge Rate to 84.8
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '84.8', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Elite Legendary - Shield Recharge Rate to 86.4
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '86.4', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Ultra Rare - Shield Recharge Rate to 88
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '88', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Exotic - Shield Recharge Rate to 89.6
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '89.6', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update


-- UPDATING Zaibatsu Heavy Industries - Guardian - Shield Recharge Rate to 91.2
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '91.2', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Zaibatsu Heavy Industries - Mythic - Shield Recharge Rate to 92.8
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '92.8', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update


-- UPDATING Zaibatsu Heavy Industries - Deus ex - Shield Recharge Rate to 96
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Shield Recharge Rate' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '96', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Zaibatsu"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Tenshi Mk1", "trait_type": "Model"}]'; -- mech model to update

/************************************
    Boston mech updates
**************************************/

-- UPDATING Boston Cybernetics - Colossal - Speed to 2,805
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '2805', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Colossal"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Rare - Speed to 2860
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '2860', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Rare"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Legendary - Speed to 2915
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '2915', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Legendary"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Elite Legendary - Speed to 2970
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '2970', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Elite Legendary"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Ultra Rare - Speed to 3025
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '3025', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Ultra Rare"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Exotic - Speed to 3080
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '3080', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Exotic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update


-- UPDATING Boston Cybernetics - Guardian - Speed to 3135
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '3135', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Guardian"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update

-- UPDATING Boston Cybernetics - Mythic - Speed to 3190
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '3190', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Mythic"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update


-- UPDATING Boston Cybernetics - Deus ex - Speed to 3300
WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'Speed' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '3300', FALSE)
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "Rarity", "value": "Deus ex"}]' -- rarity we want to update
  AND xsyn_metadata.attributes @> '[{"trait_type": "Brand", "value": "Boston Cybernetics"}]' -- mech brand to update
  AND xsyn_metadata.attributes @> '[{"value": "Law Enforcer X-1000", "trait_type": "Model"}]'; -- mech model to update
