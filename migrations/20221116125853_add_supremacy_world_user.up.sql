INSERT INTO roles (id, name, permissions)
VALUES
    ('3e05a0c2-84b0-4fa4-9626-d8020876a863', 'Supremacy World','{}');

INSERT INTO users (id, username, role_id, verified)
VALUES
    ('ba8ce250-7901-48fa-bf0c-52cd90fe139f', 'Supremacy-World', '3e05a0c2-84b0-4fa4-9626-d8020876a863', true);
