ALTER TABLE user_assets
    ADD COLUMN keywords TSVECTOR;

CREATE INDEX idx_user_assets_vec ON user_assets USING gin (keywords);

UPDATE user_assets
SET keywords =
                SETWEIGHT(TO_TSVECTOR('english', COALESCE(name, '')), 'A')
                || SETWEIGHT(TO_TSVECTOR('english', COALESCE(tier, '')), 'A')
            || SETWEIGHT(TO_TSVECTOR('english', COALESCE(asset_type, '')), 'A')
WHERE user_assets.keywords IS NULL;

CREATE OR REPLACE FUNCTION updateUserAssetsKeywords()
    RETURNS TRIGGER
AS
$updateUserAssetsKeywords$
DECLARE
    temp TSVECTOR;
BEGIN
    SELECT (SETWEIGHT(TO_TSVECTOR('english', NEW.name), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', NEW.tier), 'A') ||
            SETWEIGHT(TO_TSVECTOR('english', COALESCE(NEW.asset_type, '')), 'A'))
    INTO temp;
    IF TG_OP = 'INSERT' OR temp != OLD.keywords THEN
        UPDATE
            user_assets
        SET keywords = temp
        WHERE id = NEW.id;
    END IF;
    RETURN NULL;
END;
$updateUserAssetsKeywords$
    LANGUAGE plpgsql;

CREATE TRIGGER updateUserAssetsKeywords
    AFTER INSERT OR UPDATE
    ON user_assets
    FOR EACH ROW
EXECUTE PROCEDURE updateUserAssetsKeywords();

