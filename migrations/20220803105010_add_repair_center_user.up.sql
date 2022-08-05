-- INSERT INTO accounts (id, type, sups)
-- VALUES
--     ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'USER', 0),
--     ('5bca9b58-a71c-4134-85d4-50106a8966dc', 'USER', 0);

-- INSERT INTO roles (id, name, permissions)
-- VALUES
--     ('1db10a78-8d68-4606-b13e-f46bb84a134b', 'Repair Centre','{}'),
--     ('7841a0a3-b145-466e-bb9c-ba1327c1856f', 'Supremacy Challenge Fund','{}');
--
-- INSERT INTO users (id, username, role_id, verified, account_id)
-- VALUES
--     ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'Repair-Centre', '1db10a78-8d68-4606-b13e-f46bb84a134b', true, 'a988b1e3-5556-4cad-83bd-d61c2b149cb7'),
--     ('5bca9b58-a71c-4134-85d4-50106a8966dc', 'Supremacy-Challenge-Fund', '7841a0a3-b145-466e-bb9c-ba1327c1856f', true, '5bca9b58-a71c-4134-85d4-50106a8966dc');


INSERT INTO roles (id, name, permissions)
VALUES
    ('1db10a78-8d68-4606-b13e-f46bb84a134b', 'Repair Centre','{}'),
    ('7841a0a3-b145-466e-bb9c-ba1327c1856f', 'Supremacy Challenge Fund','{}');

INSERT INTO users (id, username, role_id, verified)
VALUES
    ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'Repair-Centre', '1db10a78-8d68-4606-b13e-f46bb84a134b', true),
    ('5bca9b58-a71c-4134-85d4-50106a8966dc', 'Supremacy-Challenge-Fund', '7841a0a3-b145-466e-bb9c-ba1327c1856f', true);
