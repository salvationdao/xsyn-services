update xsyn_store xs
set restriction  = ''
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Colossal"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Rare"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Legendary"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Elite Legendary"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Ultra Rare"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Exotic"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Guardian"}]';


update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Mythic"}]';

update xsyn_store xs
set restriction  = '' 
where xs."attributes" @> '[{"trait_type": "Rarity"}]' 
and xs."attributes" @> '[{"value": "Deus ex"}]';


update xsyn_store xs
set restriction = '' 
where xs.name IN ('Boston Cybernetics Law Enforcer X-1000 Gold ', 
 'Red Mountain Olympus Mons LY07 Gold ',
 'Zaibatsu Tenshi Mk1 Gold ');

 