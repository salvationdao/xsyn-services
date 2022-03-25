ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_type_check,
    ADD CONSTRAINT api_keys_type_check CHECK (type IN ('ADMIN','MODERATOR', 'MEMBER', 'GUEST', 'SERVER_CLIENT'));

INSERT INTO api_keys (id, user_id, type)
VALUES ('e79422b7-7bfe-4463-897b-a1d22bf2e0bc', '4fae8fdf-584f-46bb-9cb9-bb32ae20177e', 'SERVER_CLIENT');
