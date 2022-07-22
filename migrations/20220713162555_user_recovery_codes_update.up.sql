ALTER TABLE user_recovery_codes ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE users SET verified = FALSE;