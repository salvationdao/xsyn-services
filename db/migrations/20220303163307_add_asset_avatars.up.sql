BEGIN;

ALTER TABLE xsyn_store
    ADD COLUMN image_avatar TEXT NOT NULL DEFAULT '';

UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_crystal-blue_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_crystal-blue.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_crystal-blue.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Crystal Blue ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_rust-bucket_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_rust-bucket.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_rust-bucket.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Rust Bucket ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dune.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dune.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Dune ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Dynamic Yellow ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_molten_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_molten.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_molten.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Molten ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_mystermech_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_mystermech.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_mystermech.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Mystermech ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_nebula_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_nebula.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_nebula.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Nebula ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_sleek_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_sleek.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_sleek.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Sleek ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_vintage_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_vintage.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_vintage.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Vintage ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_blue-white_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_blue-white.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_blue-white.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Blue White ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_biohazard_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_biohazard.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_biohazard.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 BioHazard ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_cyber_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_cyber.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_cyber.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Cyber ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_gold.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Gold ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_light-blue-police_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_light-blue-police.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_light-blue-police.png'
    WHERE name = 'Boston Cybernetics Law Enforcer X-1000 Light Blue Police ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_vintage_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_vintage.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_vintage.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Vintage ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-white_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-white.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-white.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Red White ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-hex.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-hex.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Red Hex ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_gold.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Gold ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_desert_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_desert.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_desert.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Desert ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_navy_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_navy.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_navy.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Navy ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_nautical_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_nautical.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_nautical.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Nautical ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_military_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_military.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_military.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Military ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_irradiated_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_irradiated.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_irradiated.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Irradiated ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_evo_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_evo.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_evo.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Evo ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_beetle_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_beetle.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_beetle.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Beetle ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_villain_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_villain.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_villain.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Villain ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_green-yellow_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_green-yellow.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_green-yellow.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Green Yellow ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-blue_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-blue.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-blue.png'
    WHERE name = 'Red Mountain Olympus Mons LY07 Red Blue ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 White Gold ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_vector_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_vector.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_vector.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Vector ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_cherry-blossom_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_cherry-blossom.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_cherry-blossom.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Cherry Blossom ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_warden_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_warden.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_warden.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Warden ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gumdan_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gumdan.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gumdan.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Gumdan ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold-pattern_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold-pattern.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold-pattern.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 White Gold Pattern ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_evangelic_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_evangelic.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_evangelic.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Evangelica ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_chalky-neon_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_chalky-neon.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_chalky-neon.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Chalky Neon ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_black-digi_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_black-digi.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_black-digi.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Black Digi ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_purple-haze_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_purple-haze.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_purple-haze.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Purple Haze ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_destroyer_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_destroyer.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_destroyer.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Destroyer ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_static.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_static.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Static ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_neon_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_neon.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_neon.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Neon ';
UPDATE xsyn_store
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gold.png'
    WHERE name = 'Zaibatsu Tenshi Mk1 Gold ';


-- xsyn_metadata

ALTER TABLE xsyn_metadata
    ADD COLUMN image_avatar TEXT NOT NULL DEFAULT '';

UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_crystal-blue_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_crystal-blue.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_crystal-blue.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Crystal Blue%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_rust-bucket_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_rust-bucket.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_rust-bucket.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Rust Bucket%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dune.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dune.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Dune%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Dynamic Yellow%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_molten_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_molten.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_molten.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Molten%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_mystermech_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_mystermech.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_mystermech.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Mystermech%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_nebula_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_nebula.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_nebula.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Nebula%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_sleek_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_sleek.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_sleek.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Sleek%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_vintage_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_vintage.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_vintage.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Vintage%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_blue-white_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_blue-white.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_blue-white.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Blue White%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_biohazard_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_biohazard.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_biohazard.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 BioHazard%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_cyber_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_cyber.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_cyber.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Cyber%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_gold.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Gold%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_light-blue-police_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_light-blue-police.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_light-blue-police.png'
    WHERE name LIKE 'Boston Cybernetics Law Enforcer X-1000 Light Blue Police%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_vintage_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_vintage.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_vintage.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Vintage%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-white_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-white.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-white.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Red White%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-hex.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-hex.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Red Hex%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_gold.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Gold%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_desert_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_desert.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_desert.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Desert%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_navy_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_navy.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_navy.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Navy%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_nautical_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_nautical.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_nautical.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Nautical%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_military_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_military.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_military.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Military%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_irradiated_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_irradiated.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_irradiated.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Irradiated%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_evo_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_evo.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_evo.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Evo%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_beetle_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_beetle.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_beetle.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Beetle%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_villain_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_villain.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_villain.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Villain%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_green-yellow_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_green-yellow.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_green-yellow.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Green Yellow%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-blue_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-blue.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-blue.png'
    WHERE name LIKE 'Red Mountain Olympus Mons LY07 Red Blue%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 White Gold%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_vector_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_vector.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_vector.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Vector%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_cherry-blossom_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_cherry-blossom.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_cherry-blossom.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Cherry Blossom%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_warden_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_warden.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_warden.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Warden%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gumdan_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gumdan.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gumdan.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Gumdan%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold-pattern_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold-pattern.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold-pattern.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 White Gold Pattern%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_evangelic_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_evangelic.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_evangelic.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Evangelica%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_chalky-neon_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_chalky-neon.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_chalky-neon.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Chalky Neon%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_black-digi_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_black-digi.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_black-digi.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Black Digi%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_purple-haze_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_purple-haze.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_purple-haze.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Purple Haze%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_destroyer_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_destroyer.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_destroyer.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Destroyer%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_static.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_static.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Static%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_neon_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_neon.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_neon.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Neon%';
UPDATE xsyn_metadata
    SET image_avatar = 'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gold_avatar.png',
        animation_url = 'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gold.webm',
        image = 'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gold.png'
    WHERE name LIKE 'Zaibatsu Tenshi Mk1 Gold%';

COMMIT;
