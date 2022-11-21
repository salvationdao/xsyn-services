INSERT INTO accounts (id, type, sups)
VALUES ('1bdea1cd-43c8-44c6-bcba-42f03bf0e0c1', 'USER', 0);

INSERT INTO roles (id, name, permissions)
VALUES
    ('3e05a0c2-84b0-4fa4-9626-d8020876a863', 'Supremacy World','{}');

INSERT INTO users (id, username, role_id, verified, account_id)
VALUES
    ('ba8ce250-7901-48fa-bf0c-52cd90fe139f', 'Supremacy-World', '3e05a0c2-84b0-4fa4-9626-d8020876a863', true, '1bdea1cd-43c8-44c6-bcba-42f03bf0e0c1');
