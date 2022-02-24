--
CREATE OR REPLACE FUNCTION transactions_event() RETURNS TRIGGER AS
$noifyevent$
DECLARE
    data JSON;
BEGIN
    -- Convert the old or new row to JSON, based on the kind of action.
    -- Action = DELETE?             -> OLD row
    -- Action = INSERT or UPDATE?   -> NEW row
    IF (TG_OP = 'DELETE') THEN
        data = ROW_TO_JSON(OLD);
    ELSE
        data = ROW_TO_JSON(NEW);
    END IF;

    -- Execute pg_notify(channel, notification)
    PERFORM pg_notify('transactions_event', data::TEXT);

    -- Result is ignored since this is an AFTER trigger
    RETURN NULL;
END
$noifyevent$ LANGUAGE plpgsql;

CREATE TRIGGER transactions_event
    AFTER INSERT OR UPDATE OR DELETE
    ON transactions
    FOR EACH ROW
EXECUTE PROCEDURE transactions_event();