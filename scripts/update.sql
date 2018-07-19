ALTER TABLE users
ADD COLUMN "email_comments" Boolean NOT NULL DEFAULT TRUE;

ALTER TABLE users
ADD COLUMN "email_followers" Boolean NOT NULL DEFAULT TRUE;
