-- Finally copy the data to our new partitioned table
INSERT INTO transactions_old (id, description, transaction_reference, amount, credit, debit, created_at, "group", sub_group, service_id,
                          related_transaction_id)
SELECT id,
       description,
       transaction_reference,
       amount,
       credit,
       debit,
       created_at,
       "group",
       sub_group,
       service_id,
       related_transaction_id
FROM transactions
ON CONFLICT (id) DO NOTHING;

DROP TABLE transactions;

ALTER TABLE transactions_old
    RENAME TO transactions;

-- readd FKs
-- ALTER TABLE asset_transfer_events
--     DROP CONSTRAINT asset_transfer_events_transfer_tx_id_fkey;
--
-- ALTER TABLE asset_service_transfer_events
--     DROP CONSTRAINT asset_service_transfer_events_transfer_tx_id_fkey;
--
-- ALTER TABLE asset1155_service_transfer_events
--     DROP CONSTRAINT asset1155_service_transfer_events_transfer_tx_id_fkey;
--
-- ALTER TABLE pending_refund
--     DROP CONSTRAINT pending_refund_reversal_transaction_id_fkey,
--     DROP CONSTRAINT pending_refund_transaction_reference_fkey,
--     DROP CONSTRAINT pending_refund_withdraw_transaction_id_fkey;
