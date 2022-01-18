BEGIN;

-- Blobs
CREATE TABLE blobs (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    file_name text NOT NULL,
    mime_type text NOT NULL,
    file_size_bytes bigint NOT NULL,
    extension TEXT NOT NULL,
    file bytea NOT NULL,
    views integer NOT NULL DEFAULT 0,
    hash TEXT,
    public boolean NOT NULL DEFAULT FALSE,
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);


/******************
 *  Organisations  *
 ******************/
CREATE TABLE organisations (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid (),
    slug text UNIQUE NOT NULL,
    name text NOT NULL,
    keywords tsvector,
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- for organisation text search
CREATE INDEX idx_fts_organisation_vec ON organisations USING gin (keywords);

CREATE OR REPLACE FUNCTION updateOrganisationKeywords ()
    RETURNS TRIGGER
    AS $updateOrganisationKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.name, '')), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            organisations
        SET
            keywords = temp
        WHERE
            id = NEW.id;
    END IF;
    RETURN NULL;
    END;
$updateOrganisationKeywords$
LANGUAGE plpgsql;


/*************
 *  Factions  *
 *************/
CREATE TABLE factions (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    label text NOT NULL,
    colour text NOT NULL
);


/**********
 *  Roles  *
 **********/
CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name text UNIQUE NOT NULL,
    permissions text[] NOT NULL,
    tier integer NOT NULL DEFAULT 3, -- users can never edit another user with a tier <= to their own (SUPER_ADMIN = 1, ADMIN = 2)
    reserved boolean NOT NULL DEFAULT FALSE, -- users can never modify this row if set to true
    keywords tsvector,
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- for role text search
CREATE INDEX idx_fts_role_vec ON roles USING gin (keywords);

CREATE OR REPLACE FUNCTION updateRoleKeywords ()
    RETURNS TRIGGER
    AS $updateRoleKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.name, '')), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            roles
        SET
            keywords = temp
        WHERE
            id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateRoleKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateRoleKeywords
    AFTER INSERT OR UPDATE ON roles
    EXECUTE PROCEDURE updateRoleKeywords ();


/**********
 *  Users  *
 **********/
CREATE TABLE users (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid (),
    username text UNIQUE NOT NULL,
    role_id uuid,
    --     role_id                             UUID REFERENCES roles (id), // TODO reenable roles or remove
    avatar_id uuid REFERENCES blobs (id),
    facebook_id text UNIQUE,
    google_id text UNIQUE,
    twitch_id text UNIQUE,
    faction_id uuid REFERENCES factions (id),
    email text UNIQUE,
    first_name text DEFAULT '',
    last_name text DEFAULT '',
    verified boolean NOT NULL DEFAULT FALSE,
    old_password_required boolean NOT NULL DEFAULT TRUE, -- set to false on password reset request, set back to true on password change
    two_factor_authentication_activated boolean NOT NULL DEFAULT FALSE,
    two_factor_authentication_secret text NOT NULL DEFAULT '',
    two_factor_authentication_is_set boolean NOT NULL DEFAULT FALSE,
    sups numeric(64) NOT NULL DEFAULT 0,
    public_address text UNIQUE,
    private_address text UNIQUE,
    nonce text,
    keywords tsvector,
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT sups CHECK (sups >= 0)
);

-- for user text search
CREATE INDEX idx_fts_user_vec ON users USING gin (keywords);

CREATE OR REPLACE FUNCTION updateUserKeywords ()
    RETURNS TRIGGER
    AS $updateUserKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', NEW.first_name), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.last_name), 'A') || SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.email, '')), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.username), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            users
        SET
            keywords = temp
        WHERE
            id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateUserKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateUserKeywords
    AFTER INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE updateUserKeywords ();

CREATE TABLE user_recovery_codes (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id uuid NOT NULL REFERENCES users (id),
    recovery_code text NOT NULL,
    used_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE user_organisations (
    user_id uuid NOT NULL REFERENCES users (id),
    organisation_id uuid NOT NULL REFERENCES organisations (id),
    PRIMARY KEY (user_id, organisation_id)
);

CREATE TABLE issue_tokens (
    id uuid PRIMARY KEY NOT NULL,
    user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE password_hashes (
    user_id uuid NOT NULL REFERENCES users (id),
    password_hash text NOT NULL,
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);


/*************************
 * User Activity Tracking *     
 *************************/
CREATE TABLE user_activities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id uuid NOT NULL REFERENCES users (id),
    action text NOT NULL,
    object_id text, -- uuid
    object_slug text, -- slug/username used for links in user activity list
    object_name text, -- user friendly name for user activity list
    object_type text NOT NULL, -- enum defined in user_activities.go
    old_data json, -- old data set
    new_data json, -- new data set
    keywords tsvector,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- for user activity text search
CREATE INDEX idx_user_activities ON user_activities USING gin (keywords);

CREATE OR REPLACE FUNCTION updateUserActivityKeywords ()
    RETURNS TRIGGER
    AS $updateUserActivityKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', NEW.action), 'A') || SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.object_name, '')), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.object_type), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            user_activities
        SET
            keywords = temp
        WHERE
            id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateUserActivityKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateUserActivityKeywords
    AFTER INSERT OR UPDATE ON user_activities
    FOR EACH ROW
    EXECUTE PROCEDURE updateUserActivityKeywords ();


/*************
 *  Products  *
 *************/
CREATE TABLE products (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid (),
    slug text UNIQUE NOT NULL,
    image_id uuid REFERENCES blobs (id),
    name text NOT NULL,
    description text NOT NULL,
    keywords tsvector, -- search
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- for product text search
CREATE INDEX idx_fts_product_vec ON products USING gin (keywords);

CREATE OR REPLACE FUNCTION updateProductKeywords ()
    RETURNS TRIGGER
    AS $updateProductKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', NEW.slug), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            products
        SET
            keywords = temp
        WHERE
            id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateProductKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateProductKeywords
    AFTER INSERT OR UPDATE ON products
    FOR EACH ROW
    EXECUTE PROCEDURE updateProductKeywords ();


/********************************************
 *           xsyn_nft_metadatas              *
 * This table is the nft metadata NOT assets *
 **********************************************/
CREATE SEQUENCE IF NOT EXISTS token_id_seq;

CREATE TABLE xsyn_nft_metadata (
    token_id numeric(78, 0) PRIMARY KEY NOT NULL,
    name text NOT NULL,
    game text,
    game_object jsonb,
    description text,
    external_url text,
    image text,
    durability int NOT NULL DEFAULT 100,
    attributes jsonb,
    additional_metadata jsonb,
    keywords tsvector, -- search
    deleted_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- for xsyn_nft_metadata text search
CREATE INDEX idx_fts_xsyn_nft_metadata_vec ON xsyn_nft_metadata USING gin (keywords);

CREATE OR REPLACE FUNCTION updateXsyn_nft_metadataKeywords ()
    RETURNS TRIGGER
    AS $updateXsyn_nft_metadataKeywords$
DECLARE
    temp tsvector;
BEGIN
    SELECT
        (SETWEIGHT(TO_TSVECTOR('english', NEW.external_url), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.game), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.image), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A')) INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            xsyn_nft_metadata
        SET
            keywords = temp
        WHERE
            token_id = NEW.token_id;
    END IF;
    RETURN NULL;
END;
$updateXsyn_nft_metadataKeywords$
LANGUAGE plpgsql;

CREATE TRIGGER updateXsyn_nft_metadataKeywords
    AFTER INSERT OR UPDATE ON xsyn_nft_metadata
    FOR EACH ROW
    EXECUTE PROCEDURE updateXsyn_nft_metadataKeywords ();


/**********************************************************
 *                             Assets                      *
 * This is the table of who owns what xsync nft off chain  *
 ***********************************************************/
CREATE TABLE xsyn_assets (
    token_id numeric(78, 0) PRIMARY KEY REFERENCES xsyn_nft_metadata (token_id),
    owner_id uuid REFERENCES users (id) NOT NULL,
    frozen_at timestamptz,
    frozen_by_id uuid, -- freeze the asset if it join the game queue
    locked_by_id uuid, -- lock the asset if it is used in game
    transferred_in_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE xsyn_transaction_log (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid (),
    from_id uuid NOT NULL REFERENCES users (id),
    to_id uuid NOT NULL REFERENCES users (id),
    amount numeric(64) NOT NULL,
    status text CHECK (status IN ('pending', 'failed', 'success')) DEFAULT 'pending',
    transaction_reference text,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

COMMIT;

