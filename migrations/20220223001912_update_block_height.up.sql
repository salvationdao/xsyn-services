ALTER TABLE state ADD COLUMN latest_block_eth_testnet BIGINT;
ALTER TABLE state ADD COLUMN latest_block_bsc_testnet BIGINT;

INSERT INTO state (latest_eth_block, latest_bsc_block, latest_block_eth_testnet, latest_block_bsc_testnet, eth_to_usd, bnb_to_usd, sup_to_usd) VALUES(14256974, 15483447, 6402269, 16886589, 2589, 350, .12);