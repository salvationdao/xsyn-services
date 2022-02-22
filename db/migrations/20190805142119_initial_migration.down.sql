BEGIN;
DROP TRIGGER user_notify_event;
DROP FUNCTION user_update_event;
DROP TABLE chain_confirmations;
DROP TRIGGER trigger_check_balance;
DROP FUNCTION check_balances;
DROP VIEW account_ledgers;
DROP TABLE transactions;
DROP TABLE xsyn_assets;
DROP TABLE war_machine_ability_sups_cost;
DROP TRIGGER updateXsyn_store_name;
DROP TRIGGER updateXsyn_metadata_name;
DROP FUNCTION updateXsyn_metadata_name ();
DROP TRIGGER updateXsyn_metadataKeywords;
DROP FUNCTION updateXsyn_metadataKeywords ();
DROP TABLE xsyn_metadata;
DROP SEQUENCE IF NOT EXISTS token_id_seq;
DROP TRIGGER updatexsyn_storeKeywords;
DROP FUNCTION update_xsyn_storeKeywords ();
DROP TABLE xsyn_store;
DROP TRIGGER updateCollectionsKeywords;
DROP FUNCTION updateCollectionsKeywords ();
DROP TABLE collections;
DROP TRIGGER updateProductKeywords;
DROP FUNCTION updateProductKeywords ();
DROP TABLE products;
DROP TRIGGER updateUserActivityKeywords;
DROP FUNCTION updateUserActivityKeywords ();
DROP TABLE user_activities;
DROP TABLE password_hashes;
DROP TABLE issue_tokens;
DROP TABLE user_organisations;
DROP TABLE user_recovery_codes;
DROP TRIGGER updateUserKeywords;
DROP FUNCTION updateUserKeywords ();
DROP TABLE users;
DROP TRIGGER updateRoleKeywords;
DROP FUNCTION updateRoleKeywords ();
DROP TABLE roles;
DROP TABLE factions;
DROP FUNCTION updateOrganisationKeywords ();
DROP TABLE organisations;
DROP TABLE blobs;
DROP TABLE state;
COMMIT;
