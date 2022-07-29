INSERT INTO accounts (id, type, sups)
VALUES ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'USER', 0);

INSERT INTO roles (id, name, permissions)
VALUES ('1db10a78-8d68-4606-b13e-f46bb84a134b', 'Repair Center','{}');

INSERT INTO users (id, username, role_id, verified, account_id)
VALUES ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'Repair Center', '1db10a78-8d68-4606-b13e-f46bb84a134b', true, 'a988b1e3-5556-4cad-83bd-d61c2b149cb7');