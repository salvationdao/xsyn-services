alter table xsyn_store add column white_listed_addresses TEXT[];

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

-- uopdate whitelist mechs (gold skin)
update xsyn_store xs
set restriction = 'WHTIELIST' 
where xs.name IN ('Boston Cybernetics Law Enforcer X-1000 Gold ', 
 'Red Mountain Olympus Mons LY07 Gold ',
 'Zaibatsu Tenshi Mk1 Gold ');

