ALTER TABLE state ADD COLUMN withdraw_start_at TIMESTAMPTZ NOT NULL DEFAULT '2022-03-13T11:00:00Z';
ALTER TABLE state ADD COLUMN cliff_end_at TIMESTAMPTZ NOT NULL DEFAULT '2022-04-13T11:00:00Z';
ALTER TABLE state ADD COLUMN drip_start_at TIMESTAMPTZ NOT NULL DEFAULT '2022-04-14T11:00:00Z';


