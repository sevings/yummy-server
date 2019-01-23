
-- CREATE INDEX "index_comment_date" ---------------------------
CREATE INDEX "index_comment_date" ON "mindwell"."comments" USING btree( "created_at" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION ban_user(userName TEXT) RETURNS TEXT AS $$
    BEGIN
        UPDATE mindwell.users
        SET api_key = (
            SELECT array_to_string(array(
                SELECT substr('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789', 
                    trunc(random() * 62)::integer + 1, 1)
                FROM generate_series(1, 32)), '')
            ),
            password_hash = '', verified = false
        WHERE lower(users.name) = lower(userName);

        RETURN (SELECT email FROM mindwell.users WHERE lower(name) = lower(userName));
    END;
$$ LANGUAGE plpgsql;

ALTER TABLE entry_votes
ADD COLUMN "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL;

ALTER TABLE comment_votes
ADD COLUMN "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL;
