INSERT INTO kv (key, value) VALUES ('enable_eth_deposits', 'false') ON CONFLICT DO NOTHING;
INSERT INTO kv (key, value) VALUES ('enable_eth_withdraws', 'false') ON CONFLICT DO NOTHING;
INSERT INTO kv (key, value) VALUES ('enable_bsc_deposits', 'true') ON CONFLICT DO NOTHING;
INSERT INTO kv (key, value) VALUES ('enable_bsc_withdraws', 'true') ON CONFLICT DO NOTHING;
