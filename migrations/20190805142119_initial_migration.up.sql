BEGIN;

-- Blobs
CREATE TABLE blobs
(
    id              uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    file_name       text             NOT NULL,
    mime_type       text             NOT NULL,
    file_size_bytes bigint           NOT NULL,
    extension       TEXT             NOT NULL,
    file            bytea            NOT NULL,
    views           integer          NOT NULL DEFAULT 0,
    hash            text,
    public          BOOLEAN          NOT NULL DEFAULT FALSE,

    deleted_at      timestamptz,
    updated_at      timestamptz      NOT NULL DEFAULT NOW(),
    created_at      timestamptz      NOT NULL DEFAULT NOW()
);


/******************
*  Organisations  *
******************/

CREATE TABLE organisations
(
    id         uuid        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT UNIQUE NOT NULL,
    name       TEXT        NOT NULL,
    keywords   tsvector,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for organisation text search
CREATE INDEX idx_fts_organisation_vec ON organisations USING gin (keywords);

CREATE
OR REPLACE FUNCTION updateOrganisationKeywords ()
    RETURNS TRIGGER
    AS $updateOrganisationKeywords$
DECLARE
temp tsvector;
BEGIN
SELECT (
           setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A')
           )
INTO temp;
IF
TG_OP = 'INSERT' OR temp != OLD.keywords THEN
UPDATE
    organisations
SET keywords = temp
WHERE id = NEW.id;
END IF;
RETURN NULL;
END;

$updateOrganisationKeywords$
LANGUAGE plpgsql;

/**********
*  Roles  *
**********/

CREATE TABLE roles
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    permissions TEXT[] NOT NULL,
    tier        integer     NOT NULL DEFAULT 3,     -- users can never edit another user with a tier <= to their own (SUPER_ADMIN = 1, ADMIN = 2)
    reserved    BOOLEAN     NOT NULL DEFAULT FALSE, -- users can never modify this row if set to true
    keywords    tsvector,

    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- for role text search
CREATE INDEX idx_fts_role_vec ON roles USING gin (keywords);

CREATE
OR REPLACE FUNCTION updateRoleKeywords ()
    RETURNS TRIGGER
    AS $updateRoleKeywords$
DECLARE
temp tsvector;
BEGIN
SELECT (
           setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A')
           )
INTO temp;
IF
TG_OP = 'INSERT' OR temp != OLD.keywords THEN
UPDATE
    roles
SET keywords = temp
WHERE id = NEW.id;
END IF;
RETURN NULL;
END;

$updateRoleKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateRoleKeywords
    AFTER INSERT OR
UPDATE ON roles
    FOR EACH ROW
    EXECUTE PROCEDURE updateRoleKeywords ();

/**********
*  Users  *
**********/

CREATE TABLE users
(
    id                                  uuid        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    username                            TEXT UNIQUE NOT NULL,
    role_id                             UUID        NOT NULL REFERENCES roles (id),
    avatar_id                           uuid REFERENCES blobs (id),
    email                               TEXT UNIQUE NOT NULL,
    first_name                          TEXT        NOT NULL,
    last_name                           TEXT        NOT NULL,
    verified                            BOOLEAN     NOT NULL             DEFAULT FALSE,
    old_password_required               BOOLEAN     NOT NULL             DEFAULT TRUE, -- set to false on password reset request, set back to true on password change
    keywords                            tsvector,                                      -- search

    two_factor_authentication_activated BOOLEAN     NOT NULL             DEFAULT FALSE,
    two_factor_authentication_secret    TEXT        NOT NULL             DEFAULT '',
    two_factor_authentication_is_set    BOOLEAN     NOT NULL             DEFAULT FALSE,

    public_address                      TEXT UNIQUE,
    nonce                               TEXT,

    deleted_at                          TIMESTAMPTZ,
    updated_at                          TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at                          TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for user text search
CREATE INDEX idx_fts_user_vec ON users USING gin (keywords);

CREATE
OR REPLACE FUNCTION updateUserKeywords ()
    RETURNS TRIGGER
    AS $updateUserKeywords$
DECLARE
temp tsvector;
BEGIN
SELECT (
               setweight(to_tsvector('english', COALESCE(NEW.first_name, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.last_name, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.email, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.username, '')), 'A')
           )
INTO temp;
IF
TG_OP = 'INSERT' OR temp != OLD.keywords THEN
UPDATE
    users
SET keywords = temp
WHERE id = NEW.id;
END IF;
RETURN NULL;
END;

$updateUserKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateUserKeywords
    AFTER INSERT OR
UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE updateUserKeywords ();

CREATE TABLE user_recovery_codes
(
    id            UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users (id),
    recovery_code text        NOT NULL,
    used_at       TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

CREATE TABLE user_organisations
(
    user_id         uuid NOT NULL REFERENCES users (id),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    PRIMARY KEY (user_id, organisation_id)
);

CREATE TABLE issue_tokens
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    created_at timestamptz      NOT NULL DEFAULT NOW()
);

CREATE TABLE password_hashes
(
    user_id       UUID        NOT NULL REFERENCES users (id),
    password_hash TEXT        NOT NULL,

    deleted_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);


/*************************
* User Activity Tracking *     
*************************/

CREATE TABLE user_activities
(
    id          uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users (id),
    action      text        NOT NULL,
    object_id   text,                 -- uuid
    object_slug text,                 -- slug/username used for links in user activity list
    object_name text,                 -- user friendly name for user activity list
    object_type text        NOT NULL, -- enum defined in user_activities.go
    old_data    json,                 -- old data set
    new_data    json,                 -- new data set
    keywords    tsvector,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

-- for user activity text search
CREATE INDEX idx_user_activities ON user_activities USING gin (keywords);

CREATE
OR REPLACE FUNCTION updateUserActivityKeywords ()
    RETURNS TRIGGER
    AS $updateUserActivityKeywords$
DECLARE
temp tsvector;
BEGIN
SELECT (
               setweight(to_tsvector('english', COALESCE(NEW.action, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.object_name, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.object_type, '')), 'A')
           )
INTO temp;
IF
TG_OP = 'INSERT' OR temp != OLD.keywords THEN
UPDATE
    user_activities
SET keywords = temp
WHERE id = NEW.id;
END IF;
RETURN NULL;
END;

$updateUserActivityKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateUserActivityKeywords
    AFTER INSERT OR
UPDATE ON user_activities
    FOR EACH ROW
    EXECUTE PROCEDURE updateUserActivityKeywords ();

/*************
*  Products  *
*************/

CREATE TABLE products
(
    id          uuid        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    image_id    uuid REFERENCES blobs (id),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL,
    keywords    tsvector, -- search

    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for product text search
CREATE INDEX idx_fts_product_vec ON products USING gin (keywords);

CREATE
OR REPLACE FUNCTION updateProductKeywords ()
    RETURNS TRIGGER
    AS $updateProductKeywords$
DECLARE
temp tsvector;
BEGIN
SELECT (
               setweight(to_tsvector('english', COALESCE(NEW.slug, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
               setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'A')
           )
INTO temp;
IF
TG_OP = 'INSERT' OR temp != OLD.keywords THEN
UPDATE
    products
SET keywords = temp
WHERE id = NEW.id;
END IF;
RETURN NULL;
END;

$updateProductKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateProductKeywords
    AFTER INSERT OR
UPDATE ON products
    FOR EACH ROW
    EXECUTE PROCEDURE updateProductKeywords ();

COMMIT;