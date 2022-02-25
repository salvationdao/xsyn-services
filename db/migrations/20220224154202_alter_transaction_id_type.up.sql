-- temparary remove account ledgers view
DROP VIEW account_ledgers;


ALTER TABLE transactions
    ALTER COLUMN id TYPE TEXT;

ALTER TABLE chain_confirmations
    ALTER COLUMN tx_id TYPE TEXT;


-- put account ledgers view back
CREATE VIEW account_ledgers
            (
             account_id,
             entry_id,
             amount
                )
AS
SELECT transactions.credit,
       transactions.id,
       transactions.amount
FROM transactions
UNION ALL
SELECT transactions.debit,
       transactions.id,
       (0.0 - transactions.amount)
FROM transactions;