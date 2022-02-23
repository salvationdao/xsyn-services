BEGIN;

CREATE TABLE state
(
    latest_eth_block BIGINT,
    latest_bsc_block BIGINT,
    eth_to_usd DECIMAL,
    bnb_to_usd DECIMAL,
    sup_to_usd DECIMAL
);


-- Blobs
CREATE TABLE blobs
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid
        (
        ),
    file_name       TEXT             NOT NULL,
    mime_type       TEXT             NOT NULL,
    file_size_bytes BIGINT           NOT NULL,
    EXTENSION       TEXT             NOT NULL,
    file            BYTEA            NOT NULL,
    views           INTEGER          NOT NULL DEFAULT 0,
    HASH            TEXT,
    PUBLIC          BOOLEAN          NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW
        (
        ),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW
        (
        )
);


/******************
 *  Organisations  *
 ******************/
CREATE TABLE organisations
(
    id         UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT UNIQUE NOT NULL,
    name       TEXT        NOT NULL,
    keywords   TSVECTOR,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for organisation text search
CREATE INDEX idx_fts_organisation_vec ON organisations USING gin (keywords);

CREATE OR REPLACE FUNCTION updateOrganisationKeywords()
    RETURNS TRIGGER
AS
$updateOrganisationKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.name, '')), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            organisations
        SET keywords = temp
        WHERE id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateOrganisationKeywords$
    LANGUAGE plpgsql;


/*************
 *  Factions  *
 *************/
CREATE TABLE factions
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label              TEXT             NOT NULL,
    theme              JSONB            NOT NULL DEFAULT '{}',
    logo_blob_id       UUID             NOT NULL REFERENCES blobs (id),
    background_blob_id UUID             NOT NULL REFERENCES blobs (id),
    description        TEXT             NOT NULL DEFAULT ''
);

/**********
 *  Roles  *
 **********/
CREATE TABLE roles
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    permissions TEXT[]      NOT NULL,
    tier        INTEGER     NOT NULL DEFAULT 3,     -- users can never edit another user with a tier <= to their own (SUPER_ADMIN = 1, ADMIN = 2)
    reserved    BOOLEAN     NOT NULL DEFAULT FALSE, -- users can never modify this row if set to true
    keywords    TSVECTOR,
    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- for role text search
CREATE INDEX idx_fts_role_vec ON roles USING gin (keywords);

CREATE OR REPLACE FUNCTION updateRoleKeywords()
    RETURNS TRIGGER
AS
$updateRoleKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.name, '')), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
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
    AFTER INSERT OR UPDATE
    ON roles
EXECUTE PROCEDURE updateRoleKeywords();


/**********
 *  Users  *
 **********/
CREATE TABLE users
(
    id                                  UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    username                            TEXT UNIQUE NOT NULL,
    role_id                             UUID REFERENCES roles (id),
    avatar_id                           UUID REFERENCES blobs (id),
    facebook_id                         TEXT UNIQUE,
    google_id                           TEXT UNIQUE,
    twitch_id                           TEXT UNIQUE,
    twitter_id                          TEXT UNIQUE,
    discord_id                          TEXT UNIQUE,
    faction_id                          UUID REFERENCES factions (id),
    email                               TEXT UNIQUE,
    first_name                          TEXT                             DEFAULT '',
    last_name                           TEXT                             DEFAULT '',
    verified                            BOOLEAN     NOT NULL             DEFAULT FALSE,
    old_password_required               BOOLEAN     NOT NULL             DEFAULT TRUE, -- set to false on password reset request, set back to true on password change
    two_factor_authentication_activated BOOLEAN     NOT NULL             DEFAULT FALSE,
    two_factor_authentication_secret    TEXT        NOT NULL             DEFAULT '',
    two_factor_authentication_is_set    BOOLEAN     NOT NULL             DEFAULT FALSE,
    sups                                NUMERIC(28) NOT NULL             DEFAULT 0,    -- this check in this is in the trigger_check_balance trigger
    public_address                      TEXT UNIQUE,
    private_address                     TEXT UNIQUE,
    nonce                               TEXT,
    keywords                            TSVECTOR,
    deleted_at                          TIMESTAMPTZ,
    updated_at                          TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at                          TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for user text search
CREATE INDEX idx_fts_user_vec ON users USING gin (keywords);

CREATE OR REPLACE FUNCTION updateUserKeywords()
    RETURNS TRIGGER
AS
$updateUserKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.first_name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.last_name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.email, '')), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.username), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
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
    AFTER INSERT OR UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE updateUserKeywords();

CREATE TABLE user_recovery_codes
(
    id            UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users (id),

    recovery_code TEXT        NOT NULL,
    used_at       TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

CREATE TABLE user_organisations
(
    user_id         UUID NOT NULL REFERENCES users (id),
    organisation_id UUID NOT NULL REFERENCES organisations (id),
    PRIMARY KEY (user_id, organisation_id)
);

CREATE TABLE issue_tokens
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
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
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id),
    action      TEXT        NOT NULL,
    object_id   TEXT,                 -- uuid
    object_slug TEXT,                 -- slug/username used for links in user activity list
    object_name TEXT,                 -- user friendly name for user activity list
    object_type TEXT        NOT NULL, -- enum defined in user_activities.go
    old_data    JSON,                 -- old data set
    new_data    JSON,                 -- new data set
    keywords    TSVECTOR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- for user activity text search
CREATE INDEX idx_user_activities ON user_activities USING gin (keywords);

CREATE OR REPLACE FUNCTION updateUserActivityKeywords()
    RETURNS TRIGGER
AS
$updateUserActivityKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.action), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.object_name, '')), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.object_type), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
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
    AFTER INSERT OR UPDATE
    ON user_activities
    FOR EACH ROW
EXECUTE PROCEDURE updateUserActivityKeywords();


/*************
 *  Products  *
 *************/
CREATE TABLE products
(
    id          UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    image_id    UUID REFERENCES blobs (id),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL,
    keywords    TSVECTOR, -- search
    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);

-- for product text search
CREATE INDEX idx_fts_product_vec ON products USING gin (keywords);

CREATE OR REPLACE FUNCTION updateProductKeywords()
    RETURNS TRIGGER
AS
$updateProductKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.slug), 'A') || SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
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
    AFTER INSERT OR UPDATE
    ON products
    FOR EACH ROW
EXECUTE PROCEDURE updateProductKeywords();

-- collections
CREATE TABLE collections
(
    id           UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name         TEXT UNIQUE      NOT NULL,
    logo_blob_id UUID REFERENCES blobs (id),
    keywords     TSVECTOR, -- search
    deleted_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

-- for collection text search
CREATE INDEX idx_fts_collections_vec ON collections USING gin (keywords);

CREATE OR REPLACE FUNCTION updateCollectionsKeywords()
    RETURNS TRIGGER
AS
$updateCollectionsKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            collections
        SET keywords = temp
        WHERE id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateCollectionsKeywords$
    LANGUAGE plpgsql;

CREATE TRIGGER updateCollectionsKeywords
    AFTER INSERT OR UPDATE
    ON collections
    FOR EACH ROW
EXECUTE PROCEDURE updateCollectionsKeywords();


/*****************************************************
 *                   xsyn_store                  *
 * This table will hold the nft's on the marketplace *
 *****************************************************/
CREATE TABLE xsyn_store
(
    id                  UUID        NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_id          UUID REFERENCES factions (id),
    name                TEXT        NOT NULL,
    collection_id       UUID REFERENCES collections (id),
    description         TEXT        NOT NULL,
    image               TEXT        NOT NULL             DEFAULT '',
    animation_url       TEXT        NOT NULL             DEFAULT '',
    attributes          JSONB,
    additional_metadata JSONB,
    keywords            TSVECTOR, -- search
    usd_cent_cost       INT         NOT NULL,
    amount_sold         INT         NOT NULL             DEFAULT 0,
    amount_available    INT         NOT NULL             DEFAULT 9999999999,
    sold_after          TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    sold_before         TIMESTAMPTZ NOT NULL             DEFAULT '2999-01-01',
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    white_listed_addresses TEXT[]

);

-- for xsyn_store text search
CREATE INDEX idx_fts_xsyn_store_vec ON products USING gin (keywords);

CREATE OR REPLACE FUNCTION update_xsyn_storeKeywords()
    RETURNS TRIGGER
AS
$updatexsyn_storeKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (
                       SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') ||
                       SETWEIGHT(TO_TSVECTOR('english', NEW.image), 'A') ||
                       SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A')
               )
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            xsyn_store
        SET keywords = temp
        WHERE id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updatexsyn_storeKeywords$
    LANGUAGE plpgsql;

CREATE TRIGGER updatexsyn_storeKeywords
    AFTER INSERT OR UPDATE
    ON xsyn_store
    FOR EACH ROW
EXECUTE PROCEDURE update_xsyn_storeKeywords();


/********************************************
 *           xsyn_metadata              *
 * This table is the nft metadata NOT assets *
 **********************************************/
CREATE SEQUENCE IF NOT EXISTS token_id_seq;

CREATE TABLE xsyn_metadata
(
    token_id            NUMERIC(78, 0) PRIMARY KEY NOT NULL,
    name                TEXT                       NOT NULL,
    collection_id       UUID                       NOT NULL REFERENCES collections (id),
    game_object         JSONB,
    description         TEXT,
    external_url        TEXT,
    image               TEXT,
    animation_url       TEXT,
    durability          INT                        NOT NULL DEFAULT 100,
    attributes          JSONB,
    additional_metadata JSONB,
    keywords            TSVECTOR, -- search
    deleted_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ                NOT NULL DEFAULT NOW()
);

-- for xsyn_metadata text search
CREATE INDEX idx_fts_xsyn_metadata_vec ON xsyn_metadata USING gin (keywords);

CREATE OR REPLACE FUNCTION updateXsyn_metadataKeywords()
    RETURNS TRIGGER
AS
$updateXsyn_metadataKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.external_url), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.image), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.description), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            xsyn_metadata
        SET keywords = temp
        WHERE token_id = NEW.token_id;
    END IF;
    RETURN NULL;
END;
$updateXsyn_metadataKeywords$
    LANGUAGE plpgsql;
CREATE TRIGGER updateXsyn_metadataKeywords
    AFTER INSERT OR UPDATE
    ON xsyn_metadata
    FOR EACH ROW
EXECUTE PROCEDURE updateXsyn_metadataKeywords();

-- UPDATE NAME
CREATE OR REPLACE FUNCTION updateXsyn_metadata_name()
    RETURNS TRIGGER
AS
$updateXsyn_metadata_name$
DECLARE
    brand     TEXT DEFAULT '';
    model     TEXT DEFAULT '';
    sub_model TEXT DEFAULT '';
    name      TEXT DEFAULT '';
BEGIN
    IF NEW.attributes @> '[{"trait_type": "Brand"}]' THEN
        SELECT (SELECT CONCAT(arr.item_object ->> 'value', ' ')
                FROM JSONB_ARRAY_ELEMENTS(NEW.attributes) WITH ORDINALITY arr(item_object)
                WHERE arr.item_object ->> 'trait_type' = 'Brand')
        INTO brand;
    END IF;
    IF NEW.attributes @> '[{"trait_type": "Model"}]' THEN
        SELECT (SELECT CONCAT(arr.item_object ->> 'value', ' ')
                FROM JSONB_ARRAY_ELEMENTS(NEW.attributes) WITH ORDINALITY arr(item_object)
                WHERE arr.item_object ->> 'trait_type' = 'Model')
        INTO model;
    END IF;
    IF NEW.attributes @> '[{"trait_type": "SubModel"}]' THEN
        SELECT (SELECT CONCAT(arr.item_object ->> 'value', ' ')
                FROM JSONB_ARRAY_ELEMENTS(NEW.attributes) WITH ORDINALITY arr(item_object)
                WHERE arr.item_object ->> 'trait_type' = 'SubModel')
        INTO sub_model;
    END IF;
    IF NEW.attributes @> '[{"trait_type": "Name"}]' THEN
        SELECT (SELECT arr.item_object ->> 'value'
                FROM JSONB_ARRAY_ELEMENTS(NEW.attributes) WITH ORDINALITY arr(item_object)
                WHERE arr.item_object ->> 'trait_type' = 'Name')
        INTO name;
    END IF;
    NEW.name = CONCAT(brand, model, sub_model, name);
    RETURN NEW;
END;
$updateXsyn_metadata_name$ LANGUAGE plpgsql;

CREATE TRIGGER updateXsyn_metadata_name
    BEFORE INSERT OR UPDATE
    ON xsyn_metadata
    FOR EACH ROW
EXECUTE PROCEDURE updateXsyn_metadata_name();

CREATE TRIGGER updateXsyn_store_name
    BEFORE INSERT OR UPDATE
    ON xsyn_store
    FOR EACH ROW
EXECUTE PROCEDURE updateXsyn_metadata_name();

-- war_machine_ability_sups_cost store the sups cost for triggering the ability
CREATE TABLE war_machine_ability_sups_cost
(
    ability_token_id     NUMERIC(78, 0) NOT NULL REFERENCES xsyn_metadata (token_id),
    war_machine_token_id NUMERIC(78, 0) NOT NULL REFERENCES xsyn_metadata (token_id),
    sups_cost            TEXT           NOT NULL,
    PRIMARY KEY (ability_token_id, war_machine_token_id)
);

/**********************************************************
 *                             Assets                      *
 * This is the table of who owns what xsyn item off chain  *
 ***********************************************************/
CREATE TABLE xsyn_assets
(
    token_id          NUMERIC(78, 0) PRIMARY KEY REFERENCES xsyn_metadata (token_id),
    user_id           UUID REFERENCES users (id) NOT NULL,
    transferred_in_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
--     if any of the below are not null, the asset it locked
    frozen_by_id      UUID REFERENCES users (id),
    locked_by_id      UUID REFERENCES users (id),
    frozen_at         TIMESTAMPTZ,
    minting_signature TEXT DEFAULT '',
    tx_history JSONB DEFAULT '[]'
);

 CREATE TABLE transactions
(
    id                    SERIAL PRIMARY KEY,
    description           TEXT        NOT NULL                                                  DEFAULT '',
    transaction_reference TEXT UNIQUE NOT NULL                                                  DEFAULT '',
    amount                NUMERIC(28) NOT NULL CHECK (amount > 0.0),
    -- Every entry is a credit to one account...
    credit                UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    -- And a debit to another
    debit                 UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    -- In a paper ledger, the entry would be recorded once in each account, but
    -- that would be silly in a relational database

    status                TEXT        NOT NULL NOT NULL CHECK (status IN ('failed', 'success')) DEFAULT 'failed',
    reason                TEXT                                                                  DEFAULT '',
    -- Deletes are restricted because deleting an account with outstanding
    -- entries just doesn't make sense.  If the account's balance is nonzero,
    -- it would make assets or liabilities vanish, and even if it is zero,
    -- the account is still responsible for the nonzero balances of other
    -- accounts, so deleting it would lose important information.
    created_at            TIMESTAMPTZ NOT NULL                                                  DEFAULT NOW()
);

CREATE INDEX ON transactions (credit);
CREATE INDEX ON transactions (debit);


CREATE VIEW account_ledgers
            (
             account_id,
             entry_id,
             amount
                )
AS
SELECT transactions.credit,
       transactions.id,
       transactions.amount
FROM transactions
UNION ALL
SELECT transactions.debit,
       transactions.id,
       (0.0 - transactions.amount)
FROM transactions;

CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughFunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF NEW.debit = NEW.credit THEN
        NEW.reason = 'cannot transfer to self';
        RETURN NEW;
    END IF;
    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT NEW.debit = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (SELECT sups >= NEW.amount
                                                                  FROM users
                                                                  WHERE id = NEW.debit)
    INTO enoughFunds;
    -- if enough funds then make the updates to the user table
    IF enoughFunds THEN
        UPDATE users SET sups = sups + NEW.amount WHERE id = NEW.credit;
        UPDATE users SET sups = sups - NEW.amount WHERE id = NEW.debit;
        NEW.status = 'success';
        RETURN NEW;
    END IF;
    NEW.reason = 'insufficient funds';
    RETURN NEW;
    -- if not enough funds,
END
$check_balances$ LANGUAGE plpgsql;

-- this is the trigger before performing a entry that checks a users balance
CREATE TRIGGER trigger_check_balance
    BEFORE INSERT
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE check_balances();


/***********************************************************
 *               Waiting for confirmations table           *
 ***********************************************************/

CREATE TABLE chain_confirmations
(
    tx           TEXT PRIMARY KEY,
    tx_id        BIGINT REFERENCES transactions (id),
    block        NUMERIC(78, 0) NOT NULL,
    chain_id     NUMERIC(78, 0) NOT NULL,
    confirmed_at TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    confirmation_amount INTEGER NOT NULL DEFAULT 0
);


-- Set permissions
GRANT ALL ON transactions TO passport_tx;
GRANT ALL ON users TO passport_tx;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO passport_tx;
-- REVOKE ALL ON transactions FROM passport;
-- GRANT SELECT ON transactions TO passport;
-- GRANT UPDATE ON transactions TO passport;

/***********************************************************
 *                User listen trigger                      *
 ***********************************************************/
--
CREATE OR REPLACE FUNCTION user_update_event() RETURNS TRIGGER AS
$noifyevent$
DECLARE
    data JSON;
BEGIN
    -- Convert the old or new row to JSON, based on the kind of action.
    -- Action = DELETE?             -> OLD row
    -- Action = INSERT or UPDATE?   -> NEW row
    IF (TG_OP = 'DELETE') THEN
        data = ROW_TO_JSON(OLD);
    ELSE
        data = ROW_TO_JSON(NEW);
    END IF;

    -- Execute pg_notify(channel, notification)
    PERFORM pg_notify('user_update_event', data::TEXT);

    -- Result is ignored since this is an AFTER trigger
    RETURN NULL;
END
$noifyevent$ LANGUAGE plpgsql;

CREATE TRIGGER user_notify_event
    AFTER INSERT OR UPDATE OR DELETE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE user_update_event();

COMMIT;
