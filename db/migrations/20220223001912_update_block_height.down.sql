ALTER TABLE state ADD COLUMN latest_block_eth_testnet BIGINT DEFAULT 6402269;
ALTER TABLE state ADD COLUMN latest_block_bsc_testnet BIGINT DEFAULT 16886589;
UPDATE state SET latest_eth_block = 14256974;
UPDATE state SET latest_bsc_block = 15483447;