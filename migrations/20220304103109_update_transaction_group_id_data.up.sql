-- change group_id to group
ALTER TABLE transactions RENAME group_id TO "group";
-- add sub_group
ALTER TABLE transactions ADD COLUMN sub_group TEXT DEFAULT '';

-- update all groups that are uuids to be group battle and subgroup of uuid
UPDATE transactions
SET sub_group = "group", "group" = 'Battle'
WHERE LENGTH("group") = 36;

-- update all current battle groups to have the battle uuid in subgroup
UPDATE transactions
SET sub_group = right(left(transaction_reference, strpos(transaction_reference, '|') - 1), 36)
WHERE "group" = 'Battle';

CREATE INDEX group_idx ON transactions("group");
CREATE INDEX sub_group_idx ON transactions(sub_group);
