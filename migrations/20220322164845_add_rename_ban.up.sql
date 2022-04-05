ALTER TABLE users ADD COLUMN rename_banned bool;

ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_type_check,
    ADD CONSTRAINT api_keys_type_check CHECK (type IN ('ADMIN','MODERATOR', 'MEMBER', 'GUEST'));