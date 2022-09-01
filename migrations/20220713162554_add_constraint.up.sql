ALTER TABLE users
    ADD CONSTRAINT user_sup_check
        CHECK (
                    sups >= 0
                OR
                    id = '2fa1a63e-a4fa-4618-921f-4b4d28132069' -- this is the on chain users / seed account
            );
