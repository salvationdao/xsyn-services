
-- add restriction column
ALTER TABLE xsyn_store ADD COLUMN restriction TEXT NOT NULL NOT NULL CHECK (restriction IN ('', 'WHTIELIST', 'LOOTBOX' )) DEFAULT '';
ALTER TABLE xsyn_store ADD COLUMN white_listed_addresses TEXT[];
-- 

update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Colossal"}]';

update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Rare"}]';


update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Legendary"}]';


update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Elite Legendary"}]';


update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Ultra Rare"}]';



update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Exotic"}]';

update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Guardian"}]';


update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Mythic"}]';

update xsyn_store xs
set restriction  = 'LOOTBOX'
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Deus ex"}]';

