BEGIN;

CREATE EXTENSION IF NOT EXISTS "unaccent";

CREATE OR REPLACE FUNCTION slugify("value" TEXT)
RETURNS TEXT AS $$
  -- removes accents (diacritic signs) from a given string --
  WITH "unaccented" AS (
    SELECT unaccent("value") AS "value"
  ),
  -- lowercases the string
  "lowercase" AS (
    SELECT lower("value") AS "value"
    FROM "unaccented"
  ),
  -- remove single and double quotes
  "removed_quotes" AS (
    SELECT regexp_replace("value", '[''"]+', '', 'gi') AS "value"
    FROM "lowercase"
  ),
  -- replaces anything that's not a letter, number, hyphen('-'), or underscore('_') with a hyphen('-')
  "hyphenated" AS (
    SELECT regexp_replace("value", '[^a-z0-9\\-_]+', '-', 'gi') AS "value"
    FROM "removed_quotes"
  ),
  -- trims hyphens('-') if they exist on the head or tail of the string
  "trimmed" AS (
    SELECT regexp_replace(regexp_replace("value", '\-+$', ''), '^\-', '') AS "value"
    FROM "hyphenated"
  )
  SELECT "value" FROM "trimmed";
$$ LANGUAGE SQL STRICT IMMUTABLE;

CREATE OR REPLACE FUNCTION set_slug_from_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.slug := slugify(NEW.name);
  RETURN NEW;
END
$$;

ALTER TABLE collections
    ADD COLUMN slug TEXT UNIQUE;
    
UPDATE collections
    SET slug = id
    WHERE (slug = '') IS NOT FALSE; -- where slug is not empty or null
    
ALTER TABLE collections
    ALTER COLUMN slug SET NOT NULL;

UPDATE collections
    SET slug = slugify(name)
    WHERE slug = id::text;

CREATE TRIGGER "t_collections_insert"
BEFORE INSERT OR UPDATE ON "collections"
FOR EACH ROW
EXECUTE PROCEDURE set_slug_from_name();

COMMIT;
