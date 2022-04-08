INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit, "group", "sub_group")
SELECT t.id||CONCAT('-fix'), 'Reconcile deposits from on-chain balance', t.transaction_reference||CONCAT('-fix'), t.amount, t.debit, '2fa1a63e-a4fa-4618-921f-4b4d28132069', 'TOKEN', 'TRANSFER' FROM transactions t WHERE description ILIKE '%deposited%' AND debit = 'ebf30ca0-875b-4e84-9a78-0b3fa36a1f87';

CREATE TABLE failed_transactions (
    id TEXT PRIMARY KEY NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    failed_reference TEXT UNIQUE NOT NULL DEFAULT '',
    amount NUMERIC(28) NOT NULL CHECK (amount > 0.0),
    credit UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    debit UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    "group" TEXT DEFAULT '',
    sub_group TEXT DEFAULT '',
    service_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO failed_transactions(id, description, failed_reference, amount, credit, debit, "group", sub_group, service_id) 
SELECT t.id, t.description, t.transaction_reference, t.amount, t.credit, t.debit, t.group, t.sub_group, t.service_id FROM transactions t WHERE t.status = 'failed';

CREATE OR REPLACE FUNCTION check_balances() RETURNS TRIGGER AS
$check_balances$
DECLARE
    enoughFunds BOOLEAN DEFAULT FALSE;
BEGIN
    -- check its not a transaction to themselves
    IF NEW.debit = NEW.credit THEN
        INSERT INTO failed_transactions(id, description, failed_reference, amount, credit, debit, "group", sub_group, service_id) VALUES (NEW.id, NEW.description, NEW.transaction_reference, NEW.amount, NEW.credit, NEW.debit, NEW.group, NEW.sub_group, NEW.service_id);
        RETURN NULL;
    END IF;
    -- checks if the debtor is the on chain / off world account since that is the only account allow to go negative.
    SELECT NEW.debit = '2fa1a63e-a4fa-4618-921f-4b4d28132069' OR (SELECT sups >= NEW.amount
                                                                  FROM users
                                                                  WHERE id = NEW.debit)
    INTO enoughFunds;
    -- if enough funds then make the updates to the user table
    IF enoughFunds THEN
        UPDATE users SET sups = sups + NEW.amount WHERE id = NEW.credit;
        UPDATE users SET sups = sups - NEW.amount WHERE id = NEW.debit;
        RETURN NEW;
    ELSE
        INSERT INTO failed_transactions(id, description, failed_reference, amount, credit, debit, "group", sub_group, service_id) VALUES (NEW.id, NEW.description, NEW.transaction_reference, NEW.amount, NEW.credit, NEW.debit, NEW.group, NEW.sub_group, NEW.service_id);
        RETURN NULL;
    END IF;
    -- if not enough funds,
END
$check_balances$ 
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_balance
  ON transactions;
CREATE TRIGGER trigger_check_balance
    BEFORE INSERT
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE check_balances();

ALTER TABLE transactions DROP COLUMN status cascade;
