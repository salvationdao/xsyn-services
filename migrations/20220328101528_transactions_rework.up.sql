BEGIN;

CREATE OR REPLACE FUNCTION uppercase_group_and_sub_group() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.group := upper(NEW.group);
  NEW.sub_group := upper(NEW.sub_group);
  RETURN NEW;
END
$$;

UPDATE transactions 
    SET "group" = 'BATTLE'
WHERE "group"  = 'spoil of war';

UPDATE transactions 
    SET "group" = upper("group")
WHERE ("group" = '') = FALSE; -- where sub_group is not empty or null

UPDATE transactions
    SET sub_group = upper(sub_group)
WHERE (sub_group = '') = FALSE; -- where sub_group is not empty or null

CREATE TRIGGER "t_transactions_insert"
BEFORE INSERT OR UPDATE ON "transactions"
FOR EACH ROW
EXECUTE PROCEDURE uppercase_group_and_sub_group();

COMMIT;