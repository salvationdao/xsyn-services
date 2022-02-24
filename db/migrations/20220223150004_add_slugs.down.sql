BEGIN;

DROP TRIGGER "t_collections_insert" ON "collections";

ALTER TABLE collections
    DROP COLUMN slug;

DROP FUNCTION set_slug_from_name();

DROP FUNCTION slugify(TEXT);

DROP EXTENSION "unaccent";

COMMIT;
