UPDATE store_items SET collection_id = '8de8c66b-6146-4d61-b036-96fa721481c1', usd_cent_cost = 100000 WHERE (data -> 'blueprint_chassis' -> 'skin')::TEXT ILIKE '%slava ukraini%';
