WITH item AS (
    SELECT ('{'||pos-1||', "value"}')::TEXT[] AS path, id
    FROM xsyn_store,
         JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem ->> 'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '"Evo"', FALSE), description = 'Evo'
FROM item
WHERE item.id = xsyn_store.id
  AND xsyn_store.attributes @> '[{"trait_type": "SubModel", "value": "Eva"}]';

----
WITH item AS (
    SELECT ('{'||pos-1||', "value"}')::TEXT[] AS path, id
    FROM xsyn_store,
         JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem ->> 'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '"Gumdan"', FALSE), description = 'Gumdan'
FROM item
WHERE item.id = xsyn_store.id
  AND xsyn_store.attributes @> '[{"trait_type": "SubModel", "value": "Gundam"}]';

------

WITH item AS (
    SELECT ('{'||pos-1||', "value"}')::TEXT[] AS path, id
    FROM xsyn_store,
         JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem ->> 'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_store
SET attributes = JSONB_SET(attributes, item.path, '"Evangelica"', FALSE), description = 'Evangelica'
FROM item
WHERE item.id = xsyn_store.id
  AND xsyn_store.attributes @> '[{"trait_type": "SubModel", "value": "Evangelion"}]';

--- existing metadata

WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '"Evo"', FALSE), description = 'Evo'
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "SubModel", "value": "Eva"}]';


WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '"Evangelica"', FALSE), description = 'Evangelica'
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "SubModel", "value": "Evangelion"}]';


WITH item as (
    SELECT ('{'||pos-1||',"value"}')::text[] as path, hash
    FROM xsyn_metadata, jsonb_array_elements(attributes) WITH ORDINALITY arr(elem, pos)
    WHERE elem->>'trait_type' = 'SubModel' -- trait we want to update
)
UPDATE xsyn_metadata
SET attributes = JSONB_SET(attributes, item.path, '"Gumdan"', FALSE), description = 'Gumdan'
FROM item
WHERE item.hash = xsyn_metadata.hash
  AND xsyn_metadata.attributes @> '[{"trait_type": "SubModel", "value": "Gundan"}]';
