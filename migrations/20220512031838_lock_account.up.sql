ALTER TABLE users ADD COLUMN withdraw_lock BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN mint_lock BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN total_lock BOOLEAN NOT NULL DEFAULT false;

