ALTER TABLE collections ADD is_visible boolean default false;

INSERT INTO collections (id, name, slug, mint_contract, stake_contract) VALUES ('8de8c66b-6146-4d61-b036-96fa721481c1','Supremacy Limited Release', 'supremacy-limited-release', '0xCA949036Ad7cb19C53b54BdD7b358cD12Cf0b810', '0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95');

UPDATE collections SET is_visible = true WHERE name = 'Supremacy Genesis';
